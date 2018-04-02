# vbar

A lightweight bar written in Vala

## Features

- Blocks can be clickable
- Blocks can have drop down menus
- Support for Font Awesome
- Style anything in the panel with CSS
- Update blocks with an interval
- D-Bus integration to trigger block updates

## Screenshots

![screenshot of vbar](https://raw.githubusercontent.com/AndrewVos/vbar/master/screenshots/simple.png)

![screenshot of vbar with popup menu](https://raw.githubusercontent.com/AndrewVos/vbar/master/screenshots/popup.png)

![screenshot of vbar with transparency](https://raw.githubusercontent.com/AndrewVos/vbar/master/screenshots/transparent.png)

## Installation

### Arch Linux

AUR: [vbar-git](https://aur.archlinux.org/packages/vbar-git)

## Configuration

It's probably best to start from the example configuration:

```
mkdir -p ~/.config/vbar
cp /usr/share/doc/vbar/examples/vbar.json ~/.config/vbar
```

### Block alignment

Blocks can either be places in the `left`, `right`, or `center` of
the bar.

For example, this config has one block in the center:

```json
{
  "blocks": {
    "left": [ ],
    "center": [
      {
        "name": "title",
        "command": "Wow I Love vbar"
      }
    ],
    "right": [ ]
  }
}
```

### Static blocks

For blocks that contain text that never change.

For example, a battery icon:

```json
{
  "blocks": {
    "right": [
      {
        "name": "battery-icon",
        "text": ""
      }
    ]
  }
}
```

### Blocks with commands

If you include a `command` key, it will be
executed once when creating the block and
the text that comes back from the command
will be used as the block text.

For example, here's a block that gets the battery
percentage and displays it.

```json
{
  "blocks": {
    "right": [
      {
        "name": "battery",
        "command": "acpi | cut -d, -f2 | sed 's/ //'"
      }
    ]
  }
}
```

### Blocks with streaming commands

You may want to use the output of a command that stays
alive as block text. `vbar` can read each line output
by a command and use that as the block text.

You can do this with the `tail_command` key.

Let's say you wanted to update battery usage every five
seconds (a contrived example, I know). This is how
you can do that:

```json
{
  "blocks": {
    "right": [
      {
        "name": "battery",
        "tail_command": "while true; do acpi | cut -d, -f2 | sed 's/ //'; sleep 5; done"
      }
    ]
  }
}
```

### Blocks with commands that update on a timer

If you include a `command` and an `interval` key,
then the command will be executed on a timer and the
text that comes back from the command will be used
as the block text.

For example, here's a block that gets the battery
percentage and displays it every five seconds:

```json
{
  "blocks": {
    "right": [
      {
        "name": "battery",
        "command": "acpi | cut -d, -f2 | sed 's/ //'",
	"interval": 5
      }
    ]
  }
}
```

### Blocks with menu items

Blocks can include `menu-items` key, which
is an array of menu items to be displayed when
the block is clicked.

Each menu item can have `text` and `command`. The latter
will be executed when you click on the menu item.

Here's an example power off icon that shows an option to
Shut down when you click it.

```json
{
  "blocks": {
    "left": [
      {
        "name": "power-off-icon",
        "text": "",
        "menu-items": [
          {
            "text": "Shut down",
            "command": "systemctl poweroff"
          }
        ]
      }
    ]
  }
}
```

### Blocks that get updated after an event

Blocks can be updated from external scripts using the update
command in `vbar`:

```
vbar --update BLOCK_NAME
```

For example, let's say we want a block that displays the currently active window title. First, we add the block:

```json
{
  "blocks": {
    "center": [
      {
        "name": "title",
        "command": "xprop -id $(xprop -root _NET_ACTIVE_WINDOW | cut -d ' ' -f 5) WM_NAME | sed -e 's/.*\"\\(.*\\)\".*/\\1/'"
      }
    ]
  }
}
```

That command will be executed once, so our window title will only be the window that was active when `vbar` launched. Pretty useless.

To make the window title update as soon as you change windows, we can use `xprop` and tell `vbar` when the active window has changed.

Run the following on startup inside your window manager:

```
xprop -root -spy _NET_ACTIVE_WINDOW | while read -r LINE; do vbar --update title; done &
```

Where `title` is the name of the block you specified in `name`.

## Styles

Everything in `vbar` can be styled with css added directly to your
configuration like so:

```json
{
  "styles": {
    "panel": [
      "font-family: Hack;"
    ],
    "block": [
      "padding-top: 5px;"
    ],
    "menu": [
      "font-weight: normal;"
    ],
    "menu :hover": [
      "background-color: #232936;",
      "color: #9aa7bd;"
    ],
    "wireless": [
      "margin-right: 10px;"
    ]
  }
}
```

In the above code `panel` applies to the entire bar,
`block` applies to all blocks, `menu` applies to any
dropdown menu you may create, and `wireless` applies
to a block named `wireless.

Note that you can also do `menu :hover` to style the
menu when it is selected.

## Transparency

It is possible to use `transparent` as a colour in css
(maybe you want your wallpaper to shine through your bar).
To do this you will need a compositor, for example `compton`.
