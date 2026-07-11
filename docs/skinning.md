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
reset - Browser normalization and application layout
components - Core application styling
vars - Skin variables
skins - Visual appearance
overrides - Final user overrides

Later layers always win/trump earlier layers.

For example:

components
↓
vars
↓
skins
↓
overrides

## Application Styles

These files belong to the application and should generally not be modified.

styles/
reset.css
common.css
app.css
forms.css
autocomplete.css
config.css
context-menu.css
splitter.css
timer.css

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

skin/
    index.css

    vars.css
    splitter.css
    timer.css

    complete.css
    pb.css

    overrides.css

The only required file is:

index.css

Everything else is optional.

### index.css

This loads every part of the skin.

Example:

@import "./vars.css";
@import "./splitter.css";
@import "./timer.css";
@import "./overrides.css";
@import "./complete.css";
@import "./pb.css";

Files may be omitted if unused.

### vars.css

This defines reusable variables.

Example:

:root {

    --color-primary: #4f0f7f;

    --segment-padding: 5px;

    --gameinfo-title-size:20px;

}

These variables are referenced throughout the application.

Changing variables is the preferred way to customize a skin.

#### Available Variables
**Split List**

--splitlist-direction

Values

column
row

Example:

:root{
--splitlist-direction:row;
}

This changes the split list from vertical to horizontal.

**Game Header**

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

Example

:root{

    --gameinfo-text-align:left;

    --gameinfo-title-size:28px;

    --gameinfo-category-font-size:16px;

}

**Segment List**

--segment-border

--segment-padding

--segment-name-width

--active-segment-background

--active-segment-text-color

Example

:root{

    --segment-padding:12px;

    --segment-border:none;

    --segment-name-width:70%;

}

**Delta Column**

--split-delta-font-family

--split-delta-text-align

Example

:root{

    --split-delta-font-family:Hack;

    --split-delta-text-align:center;

}

**Comparison Column**

--split-comparison-font-family

--split-comparison-text-align

Example

:root{

    --split-comparison-text-align:left;

}

**Primary Colors**

--color-primary

Used for

Game Header
Final Segment

Example

:root{

    --color-primary:
        linear-gradient(
            to bottom,
            #0b4075,
            #082850
        );

}

### splitter.css

This file controls the appearance of the split list.

Example:

#gameInfo{

    background:#222;

}

#splitList table tr.selected{

    background:red;
    color:white;

}

Anything inside the splitter may be overridden.

### timer.css

Controls timer colors.

#### Available classes

.timer-ahead

.timer-behind

.timer-gold

Example

.timer-ahead{

    color:#00ff00;

}

.timer-behind{

    color:red;

}

PB Animation

When a Personal Best is achieved:

#gameInfo.pb

#finalSegment.pb

receive the class

pb

Example

#gameInfo.pb{

    animation:pulse 1s infinite;

}

Complete Animation

After finishing a run:

#gameInfo.complete

#finalSegment.complete

receive

complete

Example

#gameInfo.complete{

    filter:brightness(120%);

}

### overrides.css

This is the final stylesheet loaded.

Use this when:

- replacing application styles
- changing layout
- adjusting spacing
- modifying components

Anything placed here overrides every previous stylesheet.

Example

button{

    border-radius:20px;

}

.datagrid{

    border:none;

}

#### Layout Overrides

A skin is not limited to colors.

Any layout can be changed.

Example

Vertical timer

#timer-container{

    display:flex;

    flex-direction:column;

}

Horizontal split list

:root{

    --splitlist-direction:row;

}

Game title on the left

#gameInfo{

    text-align:left;

    align-items:flex-start;

}

Change table spacing

#splitList td{

    padding:15px;

}

Hide the category

#gameCategory{

    display:none;

}

Move the attempt counter

#attempts{

    left:10px;

    right:auto;

}

#### Overriding Application Components

All application components may be overridden.

Examples include:

button

input

textarea

label

.container

.container-title

.datagrid

.actions

.autocomplete

.cm-panel

.cm-item

.tooltip-bubble

.icon-btn

Example

button{

    background:#202020;

    border-radius:12px;

}

button:hover{

    background:#0055ff;

}

## Fonts

The application provides:

Techna

Hack

Monofonto

Sublima

A skin may use these directly.

Example

#gameTitle{

    font-family:Sublima;

}

Or load additional fonts.

### Available IDs

The following IDs are intended for skins.

#splitter

#splitList

#splitBody

#splitContainer

#gameInfo

#gameTitle

#gameCategory

#attempts

#timer-container

#time-container

#world-record

#finalSegment

### Available Classes

.selected

.complete

.pb

.timer-ahead

.timer-behind

.timer-gold

.splitName

.splitDelta

.splitComparison

.seg-group

.seg-group-parent

.seg-group-child

.seg-group-bottom

## Best Practices

Prefer changing variables before overriding selectors.

Good

:root{

    --segment-padding:10px;

}

Instead of

#splitList td{

    padding:10px;

}

Variables are more stable across releases.

Use overrides.css only when variables cannot achieve the desired effect.

Example: Minimal Skin
minimal/

    index.css
    vars.css

index.css

@import "./vars.css";

vars.css

:root{

    --color-primary:#202020;

    --active-segment-background:#444;

    --active-segment-text-color:white;

}
Example: Full Layout Skin
retro/

    index.css

    vars.css

    splitter.css

    timer.css

    complete.css

    pb.css

    overrides.css

This skin could:

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

without changing any application source code.

## Compatibility

Future versions of OpenSplit may (will certainly!) introduce additional variables and selectors. To maximize compatibility:

- Prefer variables over selector overrides.
- Limit overrides to elements you intend to customize.
- Avoid relying on undocumented internal DOM structure where possible.
- Place custom layout changes in overrides.css to isolate them from appearance-focused styling.
