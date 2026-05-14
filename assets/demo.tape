# vhs demo tape for Centinela README
# Render with: brew install vhs && vhs assets/demo.tape
# Output:      assets/demo.gif

Output assets/demo.gif

Set Shell "zsh"
Set FontSize 18
Set Width 1200
Set Height 700
Set Theme "Catppuccin Mocha"
Set TypingSpeed 50ms
Set PlaybackSpeed 1.0

# Pretend prompt
Type "centinela init"
Sleep 500ms
Enter
Sleep 1500ms

Type "centinela start landing-page"
Sleep 500ms
Enter
Sleep 1500ms

Type "# plan step — write docs/plans/landing-page.md + specs/landing-page.feature"
Enter
Sleep 1200ms

Type "centinela complete landing-page"
Sleep 500ms
Enter
Sleep 1500ms

Type "# code step — agent writes the implementation, hook blocks wrong-step writes"
Enter
Sleep 1200ms

Type "centinela validate"
Sleep 500ms
Enter
Sleep 2000ms

Sleep 1500ms
