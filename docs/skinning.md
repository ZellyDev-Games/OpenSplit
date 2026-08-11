# OpenSplit Skin Development Guide

## Overview

OpenSplit uses a layered CSS architecture that separates:

- Application structure
- Reusable components
- Skin variables
- Skin appearance
- Final user overrides

This allows skins to completely change the appearance of OpenSplit without modifying application source code.

A skin should primarily control:

- colors
- fonts
- spacing
- borders
- animations
- visual effects

Structural changes should be isolated to `overrides.css`.

---

# CSS Layer Order

OpenSplit defines the following CSS layer order:

```css
@layer reset, vars, components, skins, overrides;
```

Layers are applied in order.

### Layer Purpose

_reset_ - Browser normalization and application layout <br>
_components_ - Core application styling <br>
_vars_ - Skin variables <br>
_skins_ - Visual appearance <br>
_overrides_ - Final user overrides <br>

Later layers always override earlier layers.

Order:

```
reset
  ↓
components
  ↓
vars
  ↓
skins
  ↓
overrides
```

## Layer Purpose

### reset

Browser normalization and application requirements.

Contains:

- document sizing
- default element behavior
- Wails drag behavior
- selection rules
- scrollbar handling

---

### components

Application-owned component styling.

Contains:

- layouts
- forms
- buttons
- tables
- editor controls
- autocomplete
- context menus
- splitter structure
- timer structure

Skins should avoid modifying these unless necessary.

---

### vars

Skin variables.

Used for values that skins commonly customize:

- colors
- fonts
- spacing
- alignment
- sizing

Variables should be preferred over selector overrides.

---

### skins

Visual appearance.

Used for:

- colors
- gradients
- animations
- typography
- active states

---

### overrides

Final customization layer.

Use for:

- layout changes
- replacing component behavior
- unsupported customization

This layer always wins.

---

# Application Styles

Application styles are owned by OpenSplit and should generally not be modified by skins.

```
styles/
├── index.css
├── reset.css
├── common.css
├── app.css
├── forms.css
├── autocomplete.css
├── config.css
├── context-menu.css
├── editor.css
├── splitter.css
└── timer.css
```

These provide:

## reset.css

Defines:

- document sizing
- Wails drag regions
- selection behavior
- scrollbar behavior

---

## common.css

Defines:

- application fonts
- root colors
- buttons
- shared containers
- common controls

Provided fonts:

- Techna
- Hack
- Monofonto
- Sublima

---

## forms.css

Defines:

- forms
- inputs
- textareas
- tables
- datagrids
- segment editing controls

---

## autocomplete.css

Defines:

- game search autocomplete
- dropdown lists
- game result display

---

## context-menu.css

Defines:

- right-click menus
- menu items
- separators

---

## config.css

Defines configuration page layout.

Includes:

- hotkey rows
- action buttons
- option containers

---

## editor.css

Defines split editor components.

Includes:

- variable selectors
- variable rows
- editor controls

---

## splitter.css

Defines split list structure.

Includes:

- split container
- game information area
- segment table
- split columns

---

## timer.css

Defines timer structure.

Includes:

- timer placement
- world record display
- timer layout

---

# Skin Structure

A skin consists of:

```
skin/
├── index.css
├── vars.css
├── splitter.css
├── timer.css
├── complete.css
├── pb.css
└── overrides.css
```

Only `index.css` is required.

All other files are optional.

---

# index.css

`index.css` loads all skin components.

Example:

```css
@import "./vars.css";
@import "./splitter.css";
@import "./timer.css";
@import "./complete.css";
@import "./pb.css";
@import "./overrides.css";
```

Recommended order:

1. Variables
2. Appearance
3. Animations
4. Overrides

`overrides.css` should be loaded last.

---

# vars.css

Variables define the customizable parts of a skin.

Example:

```css
@layer vars {
  :root {
    --color-primary: #4f0f7f;

    --segment-padding: 5px;

    --split-delta-font-family: Monofonto;
  }
}
```

---

# Available Variables

## Primary Colors

### --color-primary

Used for:

- game header
- final segment
- completion effects
- PB effects

Example:

```css
:root {
  --color-primary: linear-gradient(to bottom, #4f0f7f, #400060);
}
```

---

# Split List

## --splitlist-direction

Controls split orientation.

Values:

```
column
row
```

Example:

```css
:root {
  --splitlist-direction: row;
}
```

---

# Game Information

## Variables

```
--gameinfo-text-align

--gameinfo-title-size
--gameinfo-title-font-weight
--gameinfo-title-padding
--gameinfo-title-line-height

--gameinfo-category-margin
--gameinfo-category-padding
--gameinfo-category-font-size
--gameinfo-category-font-weight
--gameinfo-category-line-height
```

Example:

```css
:root {
  --gameinfo-text-align: left;

  --gameinfo-title-size: 28px;

  --gameinfo-category-font-size: 16px;
}
```

---

# Segment List

Variables:

```
--active-segment-background
--active-segment-text-color

--segment-border
--segment-padding
--segment-name-width
```

Example:

```css
:root {
  --segment-padding: 12px;

  --segment-border: none;

  --segment-name-width: 70%;
}
```

---

# Delta Column

Variables:

```
--split-delta-font-family
--split-delta-text-align
```

Example:

```css
:root {
  --split-delta-font-family: Hack;

  --split-delta-text-align: center;
}
```

---

# Comparison Column

Variables:

```
--split-comparison-font-family
--split-comparison-text-align
```

Example:

```css
:root {
  --split-comparison-text-align: left;
}
```

---

# splitter.css

Controls split list appearance.

Example:

```css
#gameInfo {
  background: #222;
}

.selected {
  color: white;
}
```

Common customizations:

- game title
- category
- split rows
- active segment
- columns
- spacing

---

# timer.css

Controls timer colors and states.

Available classes:

```
.timer-ahead
.timer-behind
.timer-gold
```

Example:

```css
.timer-ahead {
  color: greenyellow;
}

.timer-behind {
  color: red;
}

.timer-gold {
  color: gold;
}
```

---

# Personal Best Animation

When a personal best occurs:

```
#gameInfo
#finalSegment
```

receive:

```
.pb
```

Example:

```css
#gameInfo.pb {
  animation: pulse 1s infinite;
}
```

PB effects belong in:

```
pb.css
```

---

# Completion Animation

When a run completes:

```
#gameInfo
#finalSegment
```

receive:

```
.complete
```

Example:

```css
#gameInfo.complete {
  filter: brightness(120%);
}
```

Completion effects belong in:

```
complete.css
```

---

# overrides.css

`overrides.css` is the final stylesheet loaded.

Use it for:

- layout changes
- unsupported customization
- replacing component behavior

Example:

```css
button {
  border-radius: 20px;
}

.datagrid {
  border: none;
}
```

---

# Layout Overrides

## Horizontal Split Layout

```css
:root {
  --splitlist-direction: row;
}
```

---

## Left-aligned Game Header

```css
#gameInfo {
  text-align: left;
}
```

---

## Hide Category

```css
#gameCategory {
  display: none;
}
```

---

## Move Attempt Counter

```css
#attempts {
  left: 10px;
  right: auto;
}
```

---

## Change Segment Spacing

```css
#splitList td {
  padding: 15px;
}
```

---

# Available IDs

The following IDs are intended for skins.

```
#splitter

#splitList

#splitContainer

#gameInfo

#gameTitle

#gameCategory

#attempts

#time-container

#world-record

#finalSegment
```

---

# Available Classes

## Split Components

```
.selected

.splitName

.splitDelta

.splitComparison
```

---

## Segment Groups

```
.seg-group

.seg-group-parent

.seg-group-child

.seg-group-bottom
```

---

## Timer States

```
.timer-ahead

.timer-behind

.timer-gold
```

---

## Effects

```
.pb

.complete
```

---

## UI Components

```
.icon-btn

.has-tooltip

.tooltip-bubble

.comparison-mode

.collapseToggle
```

---

# Fonts

OpenSplit provides:

```
Techna
Hack
Monofonto
Sublima
```

Example:

```css
#gameTitle {
  font-family: Sublima;
}
```

---

# Custom Fonts

Skins may include custom fonts.

Example:

```
skin/
├── fonts/
│   └── myfont.ttf
```

Declare:

```css
@font-face {
  font-family: "MyFont";

  src: url("./fonts/myfont.ttf") format("truetype");
}
```

Use:

```css
#time-container {
  font-family: MyFont;
}
```

---

# Rearranging CSS Elements

Skins are not limited to changing colors and fonts. Using `overrides.css`, a skin can rearrange existing application elements.

This allows a skin to create different layouts while keeping the application logic unchanged.

Common rearrangements include:

- moving the timer above or beside the split list
- changing the game information layout
- placing attempts in a different location
- changing the order of flex/grid elements
- creating horizontal or vertical layouts

Structural changes should be placed in:

```
overrides.css
```

This keeps appearance-related styles separate from layout modifications.

---

## Example: Move Timer Above Split List

The default layout places the timer within the application layout. A skin can rearrange it by changing the container layout.

Example:

```css
@layer overrides {
  #splitter {
    flex-direction: column;
  }

  #time-container {
    order: -1;
  }
}
```

This moves the timer before the split list without changing application code.

---

## Example: Place Game Information Beside Splits

A skin can create a side-by-side layout:

```css
@layer overrides {
  #gameInfo {
    width: 300px;
  }

  #splitContainer {
    flex: 1;
  }
}
```

This creates a layout with:

```
+-------------+----------------+
| Game Info   | Split List     |
|             |                |
| Timer       | Segments       |
+-------------+----------------+
```

---

When rearranging elements:

- Prefer existing flex and grid containers.
- Use documented IDs when possible.
- Keep changes inside `overrides.css`.
- Avoid depending on undocumented wrappers or generated DOM structure.

A skin may completely change the visual arrangement of OpenSplit while leaving the application components untouched.

---

# Best Practices

## Prefer variables

Good:

```css
:root {
  --segment-padding: 10px;
}
```

Avoid:

```css
#splitList td {
  padding: 10px;
}
```

---

## Keep visual changes in skins

Good:

```
splitter.css
timer.css
pb.css
complete.css
```

---

## Keep layout changes isolated

Use:

```
overrides.css
```

for:

- moving elements
- changing structure
- replacing layouts

---

## Avoid internal DOM dependencies

Prefer documented selectors.

Avoid relying on:

- generated wrappers
- temporary elements
- implementation details

---

# Example Skins

## Minimal Skin

```
minimal/
├── index.css
└── vars.css
```

index.css:

```css
@import "./vars.css";
```

vars.css:

```css
:root {
  --color-primary: #202020;
  --active-segment-background: #444;
  --active-segment-text-color: white;
}
```

---

## Full Skin

```
retro/
├── index.css
├── vars.css
├── splitter.css
├── timer.css
├── pb.css
├── complete.css
└── overrides.css
```

A full skin can:

- change colors
- replace fonts
- animate PBs
- animate completions
- change split orientation
- redesign buttons
- modify tables
- restyle menus
- customize autocomplete

without changing application code.

---

# Compatibility

Future OpenSplit versions may introduce additional variables and selectors.

To maximize compatibility:

- prefer variables over selector overrides
- keep overrides minimal
- avoid undocumented DOM structure
- isolate layout changes
- keep appearance changes in skin files
- keep structural changes in overrides.css
