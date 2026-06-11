# vhs demo tape for the Centinela README — simulated Claude Code session.
# The visual content lives in assets/demo.sh; this tape just records it.
# Render with: brew install vhs && vhs assets/demo.tape
# Output:      assets/demo.gif

Output assets/demo.gif

Set Shell "bash"
Set FontSize 16
Set Width 1200
Set Height 800
Set Padding 16
Set Theme "Catppuccin Mocha"

Hide
Type "bash assets/demo.sh"
Enter
Sleep 400ms
Show

# Script runtime is ~31s (typing + sleeps); record until it finishes.
Sleep 33s
