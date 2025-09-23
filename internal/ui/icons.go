package ui

import "fyne.io/fyne/v2"

const tutorialIconSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48">
<path fill="#FF0000" d="M43.6 12.1c-.5-2-2-3.5-4-4-3.5-1-17.6-1-17.6-1s-14 0-17.6 1c-1.9.5-3.5 2.1-4 4-1 3.5-1 10.9-1 10.9s0 7.4 1 10.9c.5 2 2.1 3.5 4 4 3.5 1 17.6 1 17.6 1s14 0 17.6-1c2-.5 3.5-2 4-4 1-3.5 1-10.9 1-10.9s0-7.4-1-10.9z"/>
<path fill="#FFFFFF" d="M19 31l12-8-12-8z"/>
</svg>`

var tutorialIcon = fyne.NewStaticResource("youtube.svg", []byte(tutorialIconSVG))
