package commands

import "strings"

var bigDigits = map[rune]string{
	'0': `
 ███ 
█   █
█   █
█   █
 ███ `,

	'1': `
  █  
 ██  
  █  
  █  
 ███ `,

	'2': `
 ███ 
    █
 ███ 
█    
█████`,

	'3': `
████ 
    █
 ███ 
    █
████ `,

	'4': `
█  █ 
█  █ 
█████
   █ 
   █ `,

	'5': `
█████
█    
████ 
    █
████ `,

	'6': `
 ███ 
█    
████ 
█   █
 ███ `,

	'7': `
█████
    █
   █ 
  █  
  █  `,

	'8': `
 ███ 
█   █
 ███ 
█   █
 ███ `,

	'9': `
 ███ 
█   █
 ████
    █
 ███ `,

	':': `
     
  ░  
     
  ░  
     `,
}

func renderBigClock(timeStr string) string {
	const glyphHeight = 5
	lines := make([]string, glyphHeight)

	for _, ch := range timeStr {
		glyph, ok := bigDigits[ch]
		if !ok {
			glyph = bigDigits['0'] // fallback to '0' or blank
		}

		glyphLines := strings.Split(strings.Trim(glyph, "\n"), "\n")
		for i := range glyphHeight {
			lines[i] += glyphLines[i] + "  " // space between characters
		}
	}

	return strings.Join(lines, "\n")
}
