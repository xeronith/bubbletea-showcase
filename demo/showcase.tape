# Bubble Tea Showcase — demo recording
# Regenerate with: make demo
Output demo/showcase.mp4

Set FontSize 15
Set FontFamily "JetBrains Mono"
Set Width 1280
Set Height 800
Set Padding 24
Set Margin 40
Set MarginFill "#0D0B14"
Set BorderRadius 12
Set WindowBar Colorful
Set WindowBarSize 36
Set Theme "Catppuccin Mocha"
Set Framerate 30

Sleep 700ms
Type "./bin/showcase"
Enter
Sleep 2s

# Kitchen: fill the form and brew a cup
Type "2"
Sleep 700ms
Up
Type@20ms "Luna"
Escape
Right
Down
Right
Down
Space
Down
Enter
Sleep 2300ms

# Gallery: fuzzy search
Type "3"
Sleep 700ms
Type "/"
Type@30ms "git"
Sleep 1600ms
Escape
Sleep 300ms

# Motion: let the spectrum breathe
Type "4"
Sleep 2000ms

# Queue: watch the concurrent jobs finish
Type "5"
Sleep 4500ms

# Back home
Type "1"
Sleep 1200ms
