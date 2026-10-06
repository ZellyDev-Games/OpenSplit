# OpenSplit Skin Development Guide

OpenSplit skins are folders of CSS files and optional assets. A skin changes the appearance of the splitter without changing application code. The current system supports CSS variables, separate appearance and override layers, vertical and horizontal splitter layouts, and a built-in skin editor with a live preview.

Use the working `default` and `HomeImprovement` skins in [OpenSplit-Skins](https://github.com/zellydev-games/opensplit-skins) as examples of the current format.

## How a skin loads

Each installed skin is a directory with an `index.css` entry point:

```text
my-skin/
├── index.css
├── vars.css          # optional
├── splitter.css      # optional
├── timer.css         # optional
├── pb.css            # optional
├── complete.css      # optional
├── overrides.css     # optional
└── images/           # optional assets
```

`index.css` is required. OpenSplit loads it after the application styles. It can import other files in the same skin using relative paths:

```css
@import "./vars.css";
@import "./splitter.css";
@import "./timer.css";
@import "./pb.css";
@import "./complete.css";
@import "./overrides.css";
```

Only import files that exist. You can also put all rules in `index.css`. Keep image and font URLs relative to the CSS file that references them, for example `url("./images/background.png")`.

## Cascade layers

OpenSplit declares this layer order:

```css
@layer reset, vars, components, skins, overrides;
```

The application puts its normalization in `reset`, default values in `vars`, and component structure in `components`. Skin variables belong in `vars`; normal visual rules belong in `skins`; custom rules that need to take precedence belong in `overrides`.

```css
/* vars.css */
@layer vars {
  :root {
    --color-primary: #4f0f7f;
    --segment-padding: 8px;
  }
}
```

```css
/* splitter.css */
@layer skins {
  #gameInfo {
    background: var(--color-primary);
  }
}
```

```css
/* overrides.css */
@layer overrides {
  #gameInfo {
    border-radius: 8px;
  }
}
```

The declared layer order matters: a rule in a later layer wins over a rule in an earlier layer, regardless of selector specificity. Keep structural or exceptional changes in `overrides` and ordinary theme styling in `skins`.

## Layouts

The splitter supports two layouts, selected with `--splitter-layout`:

```css
@layer vars {
  :root {
    --splitter-layout: vertical;
  }
}
```

- `vertical` stacks game information, the vertically scrolling segment list, and the splitter information panel (timer, comparison, and world record) from top to bottom.
- `horizontal` places splitter information beside the segment list, with segments arranged across the available width.

The user can switch the layout from the splitter context menu. The active layout is reflected by `data-layout="vertical"` or `data-layout="horizontal"` on `#splitter`. Write layout-specific rules against that attribute when a skin needs different styling in each mode. The older `--splitlist-direction` variable controls the flex direction of `#splitList`; it does not select the overall splitter layout.

```css
@layer skins {
  #splitter[data-layout="horizontal"] .splitName {
    text-align: center;
  }
}
```

The layout and segment sizing calculations use minimum widths and heights. Prefer the sizing variables below so the app can account for the skin when sizing the window. Use direct layout overrides only when a variable does not cover the change.

`--splitter-gameinfo-min-width`, `--splitter-segment-row-min-width`, and `--splitter-segment-row-min-height` are aggregate values derived from the component minimums by the application. Usually, set the component variables instead of overriding these aggregate values.

## Built-in skin variables

Override variables in `vars.css` inside `@layer vars`. These are the principal variables used by the current splitter styles and sizing logic; application versions may add more.

### Colors and shared appearance

| Variable | Purpose |
| --- | --- |
| `--color-primary` | Theme background used for game information and the final segment by the default skin. Accepts any CSS background value, including gradients. |
| `--active-segment-background` | Background of the selected segment row. |
| `--active-segment-text-color` | Text color of the selected segment row. |
| `--group-bg` | Background used for grouped segment controls. |
| `--background`, `--surface`, `--surface-dark` | Application palette values. |
| `--text-color`, `--muted-text`, `--border-color` | Common text and border colors. |
| `--accent`, `--accent-color` | Common accent colors. |

### Game information

| Variable | Purpose |
| --- | --- |
| `--gameinfo-text-align` | Alignment of game information text. |
| `--gameinfo-title-size`, `--gameinfo-title-font-weight`, `--gameinfo-title-padding`, `--gameinfo-title-line-height` | Game title typography and spacing. |
| `--gameinfo-category-margin`, `--gameinfo-category-padding`, `--gameinfo-category-font-size`, `--gameinfo-category-font-weight`, `--gameinfo-category-line-height` | Category typography and spacing. |
| `--splitter-game-title-min-width`, `--splitter-game-title-min-height` | Minimum size for the title. |
| `--splitter-game-category-min-width`, `--splitter-game-category-min-height` | Minimum size for the category. |
| `--splitter-game-variable-min-width`, `--splitter-game-variable-min-height` | Minimum size for each game variable. |
| `--splitter-attempts-min-width`, `--splitter-attempts-min-height` | Minimum size for the attempt counter. |
| `--splitter-game-info-padding-width` | Horizontal game information padding accounted for by the horizontal layout sizing. |
| `--splitter-gameinfo-min-width` | Aggregate game information minimum width. Derived from title, category, variables, and attempts. |

### Segments and comparisons

| Variable | Purpose |
| --- | --- |
| `--segment-border`, `--segment-padding`, `--segment-name-width` | Segment cell border, padding, and name column width in the default skin. |
| `--split-delta-font-family`, `--split-delta-text-align` | Delta column font and alignment. |
| `--split-comparison-font-family`, `--split-comparison-text-align` | Comparison column font and alignment. |
| `--splitter-segment-icon-min-width`, `--splitter-segment-icon-min-height` | Minimum size of an icon cell when segment icons are present. |
| `--splitter-segment-name-min-width`, `--splitter-segment-name-min-height` | Minimum size of a segment name cell. |
| `--splitter-segment-delta-min-width`, `--splitter-segment-delta-min-height` | Minimum size of a delta cell. |
| `--splitter-segment-comparison-min-width`, `--splitter-segment-comparison-min-height` | Minimum size of a comparison cell. |
| `--splitter-segment-time-min-width`, `--splitter-segment-time-min-height` | Minimum size of a segment time cell. |
| `--splitter-segment-row-min-width`, `--splitter-segment-row-min-height` | Aggregate segment row minimums derived from component minimums. |
| `--splitter-comparison-min-width`, `--splitter-comparison-min-height` | Minimum size of the comparison label. |

### Timer and world record

Timer component minimums use paired `-min-width` and `-min-height` variables:

- `--splitter-timer-sign-*`
- `--splitter-timer-hours-*`
- `--splitter-timer-minutes-*`
- `--splitter-timer-seconds-*`
- `--splitter-timer-centis-*`
- `--splitter-timer-separator-hours-minutes-*`
- `--splitter-timer-separator-minutes-seconds-*`
- `--splitter-timer-separator-seconds-centis-*`

For example, `--splitter-timer-seconds-min-width` and `--splitter-timer-seconds-min-height` set the minimum size for the seconds component. Components with no displayed value are hidden and do not take part in the timer sizing calculation.

World record component minimums also use paired `-min-width` and `-min-height` variables:

- `--splitter-world-record-label-*`, `--splitter-world-record-names-*`
- `--splitter-world-record-rt-label-*`, `--splitter-world-record-rt-time-*`, `--splitter-world-record-rt-centiseconds-*`
- `--splitter-world-record-igt-label-*`, `--splitter-world-record-igt-time-*`, `--splitter-world-record-igt-centiseconds-*`

The `*` stands for `min-width` and `min-height`; for example, `--splitter-world-record-rt-time-min-width`.

Example:

```css
@layer vars {
  :root {
    --splitter-layout: horizontal;
    --splitter-game-title-min-width: 150px;
    --splitter-segment-name-min-width: 90px;
    --splitter-segment-name-min-height: 48px;
    --splitter-timer-seconds-min-width: 42px;
    --splitter-timer-seconds-min-height: 60px;
  }
}
```

## Styling files and states

File names are a convention, not a requirement. Splitting rules by purpose makes skins easier to maintain:

- `vars.css`: variables and preferred layout.
- `splitter.css`: game information and segment list appearance.
- `timer.css`: timer and timer states.
- `pb.css`: personal best effects.
- `complete.css`: completion effects.
- `overrides.css`: layout or component changes that need the final layer.

The timer uses `.timer-ahead`, `.timer-behind`, and `.timer-gold` for split times. A completed run adds `.complete` to the game information and segment list elements. A personal best adds `.pb` to those elements. They can be styled independently or in combination:

```css
@layer skins {
  .timer-ahead { color: greenyellow; }
  .timer-behind { color: red; }
  .timer-gold { color: gold; }

  #gameInfo.complete { filter: brightness(120%); }
  #gameInfo.pb { animation: pulse 1s infinite; }
}

@keyframes pulse {
  50% { filter: brightness(150%); }
}
```

## Useful selectors

The splitter's current public-facing IDs include:

```text
#splitter                 #splitterInfo
#splitList                #splitContainer
#finalSegment             #gameInfo
#gameTitle                #gameCategory
#attempts                 #time-container
#time-sign                #time-hours
#time-sep-hm              #time-minutes
#time-sep-ms              #time-seconds
#time-sep-sc              #time-centis
#world-record
#world-record-players     #world-record-real-time
#world-record-in-game-time
```

Segment classes include `.segmentRow`, `.parentRow`, `.selected`, `.segmentIcon`, `.segment-icon` (the image), `.splitName`, `.splitDelta`, `.splitComparison`, and `.splitTime`. Grouped rows also use `.parentName`, `.parentDelta`, `.parentComparison`, and `.parentTime`; expandable groups use `.collapseToggle`. The splitter has a `.comparison-mode` label and sets `data-has-segment-icons="true"` when rows contain icons. State classes include `.complete`, `.pb`, `.timer-ahead`, `.timer-behind`, and `.timer-gold`.

These selectors describe the current splitter DOM. Prefer variables for values the system exposes and avoid relying on unrelated editor or internal wrapper markup, which can change between versions.

## Fonts and assets

OpenSplit bundles the fonts `Techna`, `Hack`, `Monofonto`, and `Sublima`. Use a bundled font by name:

```css
@layer skins {
  #gameTitle { font-family: Sublima, sans-serif; }
  #time-container { font-family: Monofonto, monospace; }
}
```

A skin can include its own fonts and images:

```text
my-skin/
├── index.css
├── fonts/myfont.ttf
└── images/background.png
```

```css
@font-face {
  font-family: "MyFont";
  src: url("./fonts/myfont.ttf") format("truetype");
}

@layer skins {
  #splitList {
    background: url("./images/background.png") center / cover no-repeat;
    font-family: "MyFont", sans-serif;
  }
}
```

## Skin editor

OpenSplit includes a skin editor for creating and editing installed skins. It presents the skin's CSS files and rules, supports adding CSS files and rules, and provides a splitter preview. Use the preview to check your styling in both layouts and across the available preview states. Changes can also be made directly in the skin folder; the active skin is watched and reloaded when its files change.

## Working examples and best practices

- Start from the `default` skin for a compact baseline, or inspect `HomeImprovement` for a skin split into component stylesheets with image assets.
- Put reusable values in `vars.css`; put ordinary appearance rules in `skins`.
- Keep `overrides.css` for structural or exceptional rules that need to win the cascade.
- Use `[data-layout="vertical"]` and `[data-layout="horizontal"]` for layout-specific styling.
- Set component minimum size variables when a skin changes widths or heights so window sizing can account for those choices.
- Prefer documented selectors and variables. If a design depends on an app structure that is not exposed, treat it as version-sensitive.
