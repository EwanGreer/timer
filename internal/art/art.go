package art

import "strings"

const DoneArt = `
██████   ██████  ███    ██ ███████ 
██   ██ ██    ██ ████   ██ ██      
██   ██ ██    ██ ██ ██  ██ █████   
██   ██ ██    ██ ██  ██ ██ ██      
██████   ██████  ██   ████ ███████ 
`

var BigDigits = map[rune]string{
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

func RenderBigClock(timeStr string) string {
	const glyphHeight = 5
	lines := make([]string, glyphHeight)

	for _, ch := range timeStr {
		glyph, ok := BigDigits[ch]
		if !ok {
			glyph = BigDigits['0'] // fallback to '0' or blank
		}

		glyphLines := strings.Split(strings.Trim(glyph, "\n"), "\n")
		for i := range glyphHeight {
			lines[i] += glyphLines[i] + "  " // space between characters
		}
	}

	return strings.Join(lines, "\n")
}
