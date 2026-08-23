#!/usr/bin/env bash

set -Eeuo pipefail

#
# OpenSplit command synchronization
#
# OpenSplit is authoritative.
#
# Synchronizes:
#
#   OpenSplit/command/
#       -> OpenSplit-FactFinder/command/
#       -> OpenSplit-RacetimeGG/command/
#
# The command package is intentionally kept as individual Go files.
#
# This script verifies:
#
#   - no duplicate command names
#   - no duplicate byte values
#   - every command value is 0..255
#   - no iota
#   - same filenames in all repositories
#   - same command definitions in all repositories
#   - all command packages build
#   - consumers contain no locally authored Command type
#   - stale command files are removed
#   - final recursive diff is clean
#
# Usage:
#
#   ./tools/synccommands.sh
#
# The default action synchronizes and then performs every verification.
#
#   ./tools/synccommands.sh --check
#
# Only verifies. Nothing is modified.
#

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
OPEN_SPLIT_ROOT="$(cd -- "${SCRIPT_DIR}/.." && pwd)"

FACTFINDER_ROOT="${OPEN_SPLIT_ROOT}/../OpenSplit-FactFinder"
RACETIMEGG_ROOT="${OPEN_SPLIT_ROOT}/../OpenSplit-RacetimeGG"

OPEN_SPLIT_COMMAND="${OPEN_SPLIT_ROOT}/command"
FACTFINDER_COMMAND="${FACTFINDER_ROOT}/command"
RACETIMEGG_COMMAND="${RACETIMEGG_ROOT}/command"

CHECK_ONLY=false

case "${1:-}" in
    "")
        ;;
    --check)
        CHECK_ONLY=true
        ;;
    --help|-h)
        cat <<EOF
Usage:
    $(basename "$0")          Synchronize and verify command packages
    $(basename "$0") --check  Verify only; do not modify files
EOF
        exit 0
        ;;
    *)
        echo "ERROR: unknown option: $1" >&2
        exit 2
        ;;
esac


###############################################################################
# Helpers
###############################################################################

die() {
    echo
    echo "ERROR: $*" >&2
    exit 1
}

info() {
    echo "==> $*"
}

ok() {
    echo "    OK: $*"
}

require_directory() {
    local path="$1"
    local description="$2"

    [[ -d "$path" ]] ||
        die "${description} does not exist: ${path}"
}

require_file() {
    local path="$1"
    local description="$2"

    [[ -f "$path" ]] ||
        die "${description} does not exist: ${path}"
}


###############################################################################
# Repository validation
###############################################################################

info "Checking repositories"

require_directory \
    "$OPEN_SPLIT_ROOT" \
    "OpenSplit repository"

require_directory \
    "$FACTFINDER_ROOT" \
    "FactFinder repository"

require_directory \
    "$RACETIMEGG_ROOT" \
    "OpenSplit-RacetimeGG repository"

require_directory \
    "$OPEN_SPLIT_COMMAND" \
    "OpenSplit command directory"

require_directory \
    "$FACTFINDER_COMMAND" \
    "FactFinder command directory"

require_directory \
    "$RACETIMEGG_COMMAND" \
    "OpenSplit-RacetimeGG command directory"

ok "repositories found"


###############################################################################
# Find command source files
###############################################################################

command_files() {
    local directory="$1"

    find "$directory" \
        -maxdepth 1 \
        -type f \
        -name '*.go' \
        -printf '%f\n' |
        sort
}


###############################################################################
# Validate filenames
#
# This is deliberately done before copying anything.
#
# After synchronization the three sets must be identical.
###############################################################################

validate_filenames() {
    local name="$1"
    local directory="$2"

    local files

    files="$(command_files "$directory")"

    [[ -n "$files" ]] ||
        die "${name}: command directory contains no Go files"

    printf '%s\n' "$files"
}


info "Checking command filenames"

OPEN_SPLIT_FILES="$(validate_filenames \
    "OpenSplit" \
    "$OPEN_SPLIT_COMMAND")"

FACTFINDER_FILES="$(validate_filenames \
    "FactFinder" \
    "$FACTFINDER_COMMAND")"

RACETIMEGG_FILES="$(validate_filenames \
    "OpenSplit-RacetimeGG" \
    "$RACETIMEGG_COMMAND")"

if [[ "$CHECK_ONLY" == false ]]; then
    :
fi

ok "command filename lists collected"


###############################################################################
# Validate authoritative command definitions
#
# We deliberately require the following form:
#
#     COMMAND_NAME Command = 123
#
# rather than:
#
#     COMMAND_NAME Command = iota
#
# or:
#
#     COMMAND_NAME
#
# This makes the wire value explicit and reviewable.
###############################################################################

validate_source_definitions() {
    local directory="$1"
    local name="$2"

    info "Validating ${name} command definitions"

    local duplicate_names
    local duplicate_values
    local invalid_values
    local iota_definitions
    local malformed_definitions

    #
    # No iota anywhere in the command package.
    #
    iota_definitions="$(
        grep -REn \
            --include='*.go' \
            '\biota\b' \
            "$directory" ||
        true
    )"

    if [[ -n "$iota_definitions" ]]; then
        echo "$iota_definitions" >&2
        die "${name}: command package contains iota"
    fi

    #
    # Extract command definitions.
    #
    #
    # Expected:
    #
    #     NAME Command = NUMBER
    #
    # The command files are intentionally kept simple so that the wire
    # definition remains obvious to humans reviewing a diff.
    #
    local definitions

    definitions="$(
        grep -REh \
            --include='*.go' \
            '^[[:space:]]*[A-Z][A-Za-z0-9_]*[[:space:]]+Command[[:space:]]*=[[:space:]]*[0-9]+([[:space:]]*//.*)?$' \
            "$directory" |
        sed -E \
            's/^[[:space:]]*([A-Z][A-Za-z0-9_]*)[[:space:]]+Command[[:space:]]*=[[:space:]]*([0-9]+).*/\1 \2/' |
        sort
    )"

    [[ -n "$definitions" ]] ||
        die "${name}: no explicit command definitions found"

    #
    # Detect Command declarations which aren't explicit numeric values.
    #
    malformed_definitions="$(
        grep -REn \
            --include='*.go' \
            '^[[:space:]]*[A-Z][A-Za-z0-9_]*[[:space:]]+Command[[:space:]]*=' \
            "$directory" |
        grep -Ev \
            'Command[[:space:]]*=[[:space:]]*[0-9]+([[:space:]]*//.*)?$' ||
        true
    )"

    if [[ -n "$malformed_definitions" ]]; then
        echo "$malformed_definitions" >&2
        die "${name}: malformed command definition; expected explicit numeric value"
    fi

    #
    # Duplicate command names.
    #
    duplicate_names="$(
        printf '%s\n' "$definitions" |
        awk '{ print $1 }' |
        sort |
        uniq -d
    )"

    if [[ -n "$duplicate_names" ]]; then
        echo "$duplicate_names" >&2
        die "${name}: duplicate command name"
    fi

    #
    # Duplicate byte values.
    #
    duplicate_values="$(
        printf '%s\n' "$definitions" |
        awk '{ print $2 }' |
        sort -n |
        uniq -d
    )"

    if [[ -n "$duplicate_values" ]]; then
        echo "$duplicate_values" >&2
        die "${name}: duplicate command byte value"
    fi

    #
    # Every value must fit in the protocol command byte.
    #
    invalid_values="$(
        printf '%s\n' "$definitions" |
        awk '
            {
                value = $2

                if (value < 0 || value > 255) {
                    print $1 " " value
                }
            }
        '
    )"

    if [[ -n "$invalid_values" ]]; then
        echo "$invalid_values" >&2
        die "${name}: command value outside 0..255"
    fi

    local count
    count="$(
        printf '%s\n' "$definitions" |
        wc -l
    )"

    ok "${name}: ${count} command definitions validated"
}


###############################################################################
# Validate authoritative OpenSplit
###############################################################################

validate_source_definitions \
    "$OPEN_SPLIT_COMMAND" \
    "OpenSplit"


###############################################################################
# Synchronization
###############################################################################

sync_directory() {
    local source="$1"
    local destination="$2"
    local name="$3"

    info "Synchronizing ${name}"

    mkdir -p "$destination"

    #
    # Remove stale .go files from the consumer.
    #
    while IFS= read -r file; do
        [[ -n "$file" ]] || continue

        if [[ ! -f "${source}/${file}" ]]; then
            echo "    Removing stale ${name}/command/${file}"

            rm -f "${destination}/${file}"
        fi
    done < <(command_files "$destination")

    #
    # Copy every authoritative file.
    #
    while IFS= read -r file; do
        [[ -n "$file" ]] || continue

        cp \
            "${source}/${file}" \
            "${destination}/${file}"
    done < <(command_files "$source")

    ok "${name}: synchronized"
}


if [[ "$CHECK_ONLY" == false ]]; then
    sync_directory \
        "$OPEN_SPLIT_COMMAND" \
        "$FACTFINDER_COMMAND" \
        "FactFinder"

    sync_directory \
        "$OPEN_SPLIT_COMMAND" \
        "$RACETIMEGG_COMMAND" \
        "OpenSplit-RacetimeGG"
else
    info "Check-only mode: no files will be modified"
fi


###############################################################################
# Verify filenames
###############################################################################

info "Verifying command filenames"

verify_filenames() {
    local source="$1"
    local destination="$2"
    local name="$3"

    if ! diff -u \
        <(command_files "$source") \
        <(command_files "$destination"); then

        die "${name}: command filenames differ from OpenSplit"
    fi

    ok "${name}: filenames match OpenSplit"
}

verify_filenames \
    "$OPEN_SPLIT_COMMAND" \
    "$FACTFINDER_COMMAND" \
    "FactFinder"

verify_filenames \
    "$OPEN_SPLIT_COMMAND" \
    "$RACETIMEGG_COMMAND" \
    "OpenSplit-RacetimeGG"


###############################################################################
# Verify consumers do not contain locally authored OpenSplit commands
#
# Other packages are allowed to define their own types named Command.
#
# For example, QUSB2SNES has its own:
#
#     type Command int
#
# That is unrelated to the OpenSplit wire protocol and is valid.
#
# What consumers must not do is independently define one of the OpenSplit
# protocol commands outside the synchronized command package.
###############################################################################

info "Checking consumers for duplicate OpenSplit command definitions"

verify_no_local_opensplit_commands() {
    local root="$1"
    local command_directory="$2"
    local name="$3"

    local definitions=""
    local matches=""

    #
    # Build the authoritative list of command names.
    #
    while IFS= read -r line; do
        [[ -n "$line" ]] || continue

        local command_name
        command_name="$(awk '{print $1}' <<< "$line")"

        #
        # Search Go source outside command/ for declarations of the form:
        #
        #     SPLIT Command = ...
        #
        # We intentionally require the OpenSplit Command type here. A type
        # named Command belonging to another protocol does not match.
        #
        local found
        found="$(
            find "$root" \
                -type f \
                -name '*.go' \
                -not -path "${command_directory}/*" \
                -print0 |
            xargs -0 grep -nE \
                "^[[:space:]]*${command_name}[[:space:]]+Command[[:space:]]*=" \
            || true
        )"

        if [[ -n "$found" ]]; then
            matches+="${found}"$'\n'
        fi
    done < <(
        grep -REh \
            --include='*.go' \
            '^[[:space:]]*[A-Z][A-Za-z0-9_]*[[:space:]]+Command[[:space:]]*=[[:space:]]*[0-9]+([[:space:]]*//.*)?$' \
            "$OPEN_SPLIT_COMMAND" |
        sed -E \
            's/^[[:space:]]*([A-Z][A-Za-z0-9_]*)[[:space:]]+Command[[:space:]]*=[[:space:]]*([0-9]+).*/\1 \2/' |
        sort
    )

    if [[ -n "$matches" ]]; then
        printf '%s' "$matches" >&2
        die "${name}: locally authored OpenSplit command found outside command/"
    fi

    ok "${name}: no locally authored OpenSplit commands"
}

verify_no_local_opensplit_commands \
    "$FACTFINDER_ROOT" \
    "$FACTFINDER_COMMAND" \
    "FactFinder"

verify_no_local_opensplit_commands \
    "$RACETIMEGG_ROOT" \
    "$RACETIMEGG_COMMAND" \
    "OpenSplit-RacetimeGG"


###############################################################################
# Verify command definitions
#
# Because OpenSplit is authoritative, the consumer definitions must be
# byte-for-byte identical, not merely semantically equivalent.
###############################################################################

info "Verifying command definitions"

verify_definitions() {
    local source="$1"
    local destination="$2"
    local name="$3"

    if ! diff -ru \
        "$source" \
        "$destination"; then

        die "${name}: command definitions differ from OpenSplit"
    fi

    ok "${name}: command definitions match OpenSplit"
}

verify_definitions \
    "$OPEN_SPLIT_COMMAND" \
    "$FACTFINDER_COMMAND" \
    "FactFinder"

verify_definitions \
    "$OPEN_SPLIT_COMMAND" \
    "$RACETIMEGG_COMMAND" \
    "OpenSplit-RacetimeGG"


###############################################################################
# Validate consumers
#
# Run the same definition checks after synchronization. This ensures that
# synchronization itself did not introduce an invalid definition.
###############################################################################

validate_source_definitions \
    "$FACTFINDER_COMMAND" \
    "FactFinder"

validate_source_definitions \
    "$RACETIMEGG_COMMAND" \
    "OpenSplit-RacetimeGG"


###############################################################################
# Verify command packages build
###############################################################################

info "Building command packages"

(
    cd "$OPEN_SPLIT_ROOT"

    go test ./command
)

ok "OpenSplit command package builds"

(
    cd "$FACTFINDER_ROOT"

    go test ./command
)

ok "FactFinder command package builds"

(
    cd "$RACETIMEGG_ROOT"

    go test ./command
)

ok "OpenSplit-RacetimeGG command package builds"


###############################################################################
# Final recursive diff
#
# This is deliberately the last check.
#
# If the command directories differ for ANY reason, synchronization has
# failed.
###############################################################################

info "Running final diff verification"

verify_final_diff() {
    local source="$1"
    local destination="$2"
    local name="$3"

    if ! diff -ru \
        "$source" \
        "$destination"; then

        die "${name}: final diff -ru is not clean"
    fi

    ok "${name}: final diff -ru is clean"
}

verify_final_diff \
    "$OPEN_SPLIT_COMMAND" \
    "$FACTFINDER_COMMAND" \
    "FactFinder"

verify_final_diff \
    "$OPEN_SPLIT_COMMAND" \
    "$RACETIMEGG_COMMAND" \
    "OpenSplit-RacetimeGG"


###############################################################################
# Done
###############################################################################

echo
echo "============================================================"
echo "Command synchronization successful"
echo "============================================================"
echo
echo "Authoritative:"
echo "  ${OPEN_SPLIT_COMMAND}"
echo
echo "Synchronized:"
echo "  ${FACTFINDER_COMMAND}"
echo "  ${RACETIMEGG_COMMAND}"
echo
echo "Verified:"
echo "  - no duplicate command names"
echo "  - no duplicate byte values"
echo "  - every value is 0..255"
echo "  - no iota"
echo "  - same filenames"
echo "  - same command definitions"
echo "  - all command packages build"
echo "  - consumers contain no local OpenSplit Command type"
echo "  - stale command files removed"
echo "  - final diff -ru is clean"
echo
