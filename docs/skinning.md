# OpenSplit Skin Development Guide

## Overview

OpenSplit uses a layered CSS architecture that cleanly separates:

- Application components
- Default layout
- Theme variables
- Skin appearance
- User overrides

This allows a skin to completely change the appearance without modifying the application itself.

## CSS Layer Order

The application defines the following layer order:

@layer reset, components, vars, skins, overrides;

Layers are applied in order.

### Layer Purpose

_reset_ - Browser normalization and application layout <br>
_components_ - Core application styling <br>
_vars_ - Skin variables <br>
_skins_ - Visual appearance <br>
_overrides_ - Final user overrides <br>

Later layers always win/trump earlier layers.

For example:

components <br>
↓ <br>
vars <br>
↓ <br>
skins <br>
↓ <br>
overrides <br>

## Application Styles

These files belong to the application and should generally not be modified.

```bash
├── styles/
│   ├── reset.css
│   ├── common.css
│   ├── app.css
│   ├── forms.css
│   ├── autocomplete.css
│   ├── config.css
│   ├── context-menu.css
│   ├── splitter.css
│   ├── timer.css
```

These define:

- window layout
- reusable controls
- forms
- autocomplete
- context menus
- splitter layout
- timer layout

A skin can override nearly all of these.

## Skin Structure

A skin consists of:

```bash
├── skin/
│   ├── index.css
│   ├── vars.css
│   ├── splitter.css
│   ├── timer.css
│   ├── complete.css
│   ├── pb.css
│   ├── overrides.css
```

The only required file is **index.css**; everything else is optional.

### index.css

This loads every part of the skin.

Example:

```
@import "./vars.css";
@import "./splitter.css";
@import "./timer.css";
@import "./overrides.css";
@import "./complete.css";
@import "./pb.css";
```

Files may be omitted if unused.

### vars.css

This defines reusable variables.

Example:

```
:root {
    --color-primary: #4f0f7f;
    --segment-padding: 5px;
    --gameinfo-title-size:20px;
}
```

These variables are referenced throughout the application.

Changing variables is the preferred way to customize a skin.

#### Available Variables

**Split List** <br>

--splitlist-direction

Values can be either _column_ or _row_.

Example:

```
:root{
--splitlist-direction:row;
}
```

This changes the split list from vertical (columns) to horizontal (rows).

**Game Header** <br>
--gameinfo-text-align <br>
--gameinfo-title-size <br>
--gameinfo-title-font-weight <br>
--gameinfo-title-padding <br>
--gameinfo-title-line-height <br>
--gameinfo-category-margin <br>
--gameinfo-category-padding <br>
--gameinfo-category-font-size <br>
--gameinfo-category-font-weight <br>
--gameinfo-category-line-height <br>

Example:

```
:root{
    --gameinfo-text-align:left;
    --gameinfo-title-size:28px;
    --gameinfo-category-font-size:16px;
}
```

**Segment List** <br>
--segment-border <br>
--segment-padding <br>
--segment-name-width <br>
--active-segment-background <br>
--active-segment-text-color <br>

Example:

```
:root{
    --segment-padding:12px;
    --segment-border:none;
    --segment-name-width:70%;
}
```

**Delta Column** <br>
--split-delta-font-family <br>
--split-delta-text-align <br>

Example:

```
:root{
    --split-delta-font-family:Hack;
    --split-delta-text-align:center;
}
```

**Comparison Column** <br>
--split-comparison-font-family <br>
--split-comparison-text-align <br>

Example:

```
:root{
    --split-comparison-text-align:left;
}
```

**Primary Colors** <br>
--color-primary <br>

Used for:

- Game Header
- Final Segment

Example:

```
:root{
    --color-primary:
        linear-gradient(
            to bottom,
            #0b4075,
            #082850
        );
}
```

### splitter.css

This file controls the appearance of the split list.

Example:

```
#gameInfo{
    background:#222;
}

#splitList table tr.selected{
    background:red;
    color:white;
}
```

Anything inside the splitter may be overridden.

### timer.css

Controls timer colors.

#### Available classes

.timer-ahead <br>
.timer-behind <br>
.timer-gold <br>

Example:

```
.timer-ahead{
    color:#00ff00;
}

.timer-behind{
    color:red;
}

.timer-gold{
    color:#D4AF37;
)
```

#### PB Animation

When a personal best (PB) is achieved:

#gameInfo.pb <br>
#finalSegment.pb <br>

receive the class `pb`.

Example:

```
#gameInfo.pb{
    animation:pulse 1s infinite;
}
```

#### Complete Animation

After completing a run:

#gameInfo.complete <br>
#finalSegment.complete <br>

receive the class `complete`

Example:

```
#gameInfo.complete{
    filter:brightness(120%);
}
```

### overrides.css

This is the final stylesheet loaded and thereby trumps all other stylesheets.

Use this when:

- replacing application styles
- changing layout
- adjusting spacing
- modifying components

Again: **anything placed here overrides every previous stylesheet**.

Example:

```
button{
    border-radius:20px;
}

.datagrid{
    border:none;
}
```

#### Layout Overrides

A skin is not limited to colors. Any layout can be changed.

Some examples include:

**Vertical timer**

```
#timer-container{
    display:flex;
    flex-direction:column;
}
```

**Horizontal split list**

```
:root{
    --splitlist-direction:row;
}
```

**Game title on the left**

```
#gameInfo{
    text-align:left;
    align-items:flex-start;
}
```

**Change table spacing**

```
#splitList td{
    padding:15px;
}
```

**Hide the category**

```
#gameCategory{
    display:none;
}
```

**Move the attempt counter**

```
#attempts{
    left:10px;
    right:auto;
}
```

#### Overriding Application Components

All application components may be overridden. To name several:

- button
- input
- textarea
- label
- .container
- .container-title
- .datagrid
- .actions
- .autocomplete
- .cm-panel
- .cm-item
- .tooltip-bubble
- .icon-btn

Some more specific examples:

```
button{
    background:#202020;
    border-radius:12px;
}

button:hover{
    background:#0055ff;
}
```

## Fonts

The application provides:

- Techna
- Hack
- Monofonto
- Sublima

A skin may use these directly.

Example:

```
#gameTitle{
    font-family:Sublima;
}
```

### Locally installed fonts

You may also use any installed fonts to render within OpenSplit. If you had a font called Serpentine installed locally, an example might read:

```
#time-container {
    font-family: "Serpentine", sans-serif;
    font-size: 50px;
    }
```

It is possible to use a downloaded font present in a `fonts/` folder found within the `skin/` folder, too. You would need to declare the skin, similar to the below:

```
src: url("./fonts/<filename>.ttf") format("<type>");
```

## Available IDs

The following IDs are intended for skins.

- #splitter
- #splitList
- #splitBody
- #splitContainer
- #gameInfo
- #gameTitle
- #gameCategory
- #attempts
- #timer-container
- #time-container
- #world-record
- #finalSegment

## Available Classes

- .selected
- .complete
- .pb
- .timer-ahead
- .timer-behind
- .timer-gold
- .splitName
- .splitDelta
- .splitComparison
- .seg-group
- .seg-group-parent
- .seg-group-child
- .seg-group-bottom

## Best Practices

Prefer changing variables before overriding selectors.

Good

```
:root{
    --segment-padding:10px;
}
```

Instead of

```
#splitList td{
    padding:10px;
}
```

Variables are more stable across releases.

Use overrides.css only when variables cannot achieve the desired effect.

Some examples of skins:

**Minimal Skin** <br>

```bash
├── minimal/
│   ├── index.css
│   ├── vars.css
```

`index.css` could contain:

```
@import "./vars.css";
```

`vars.css` could contain:

```
:root{
    --color-primary:#202020;
    --active-segment-background:#444;
    --active-segment-text-color:white;
}
```

**Full Layout Skin** <br>

```bash
├── retro/
│   ├── index.css
│   ├── vars.css
│   ├── splitter.css
│   ├── timer.css
│   ├── complete.css
│   ├── pb.css
│   ├── overrides.css
```

Without changing any application source code, this skin could:

- move the timer
- make horizontal splits
- use bitmap fonts
- replace colors
- add animations
- redesign buttons
- customize tables
- restyle context menus
- modify autocomplete
- replace every application component

## Compatibility

Future versions of OpenSplit may (will certainly!) introduce additional variables and selectors. To maximize compatibility:

- Prefer variables over selector overrides.
- Limit overrides to elements you intend to customize.
- Avoid relying on undocumented internal DOM structure where possible.
- Place custom layout changes in overrides.css to isolate them from appearance-focused styling.
