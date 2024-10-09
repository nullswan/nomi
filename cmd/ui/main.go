package main

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TODO: Multi-line input
// TODO: Copyable view

const longText = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Aliquam consectetur ornare lobortis. Donec a lectus vitae enim sollicitudin fermentum. Quisque fringilla sapien vitae arcu volutpat, in lacinia augue iaculis. Aenean ac rutrum elit. In hac habitasse platea dictumst. Donec pellentesque justo porttitor tincidunt volutpat. Morbi faucibus, erat eu porta feugiat, tellus orci rutrum lectus, a maximus velit lectus interdum odio. Proin urna augue, egestas ut lectus sed, fringilla lacinia turpis. Integer porttitor quam vel felis ultricies euismod. Vivamus elementum mauris eu aliquam consequat. Nullam iaculis aliquet purus, nec mollis nisl accumsan non.

Maecenas ac suscipit felis. Nulla non justo felis. Pellentesque euismod magna vitae molestie tempus. Integer dictum enim sit amet scelerisque fringilla. Nam auctor justo nec odio faucibus, eget consequat orci tempor. Mauris efficitur arcu non risus laoreet commodo. Interdum et malesuada fames ac ante ipsum primis in faucibus. Vestibulum nisl tellus, molestie eget nisl sit amet, auctor tristique libero. Nulla magna tellus, ultricies sed consectetur quis, convallis quis orci. Sed elementum dictum sem, ut eleifend massa convallis eu. Etiam vitae nisl id mi malesuada fringilla nec in purus. Sed mi est, pulvinar non augue quis, convallis cursus ligula. Etiam egestas eros quis pharetra ullamcorper. Mauris consequat auctor aliquam. Ut luctus vel eros at tempus. Fusce a neque at arcu tincidunt vehicula non nec metus.

Nulla suscipit aliquet lorem quis cursus. Praesent convallis eleifend risus, eget cursus sapien aliquam sit amet. Quisque consequat felis sem, nec ultrices justo gravida non. Vestibulum in mauris a felis pellentesque placerat ut et enim. Nunc ut velit sed mi porttitor maximus sit amet non urna. Donec eu odio nunc. Aenean faucibus lobortis erat, in tempus ligula tempor ac. Integer iaculis tincidunt augue, vitae egestas ante fermentum a. Aliquam commodo ut ipsum eget posuere. Duis sit amet sem et neque vulputate malesuada. Donec ornare diam risus, nec ultricies leo finibus in. Nunc orci sapien, auctor et nisi ac, volutpat aliquam tellus. Maecenas venenatis venenatis libero, id aliquam tellus hendrerit et. Nulla ac nibh mattis, tempus nisl ut, lobortis lorem. Sed sit amet interdum nisi.

Etiam scelerisque justo sit amet urna vestibulum, sit amet vulputate mauris mattis. Etiam dolor justo, faucibus in elementum ut, fermentum at est. Proin nisl nibh, interdum ac eros eu, venenatis cursus arcu. Morbi semper ornare augue, at faucibus ex aliquet ac. Donec nec venenatis eros. Pellentesque aliquet fringilla lorem vitae sollicitudin. Aenean porttitor, diam vel gravida tincidunt, nibh lectus venenatis elit, sed posuere velit mauris eget erat.

Nunc accumsan condimentum turpis, in ullamcorper dui finibus non. Mauris feugiat metus ut leo blandit consectetur sed eu metus. Donec mi arcu, ultricies in ultrices ac, malesuada in nulla. Aliquam sagittis, ante ac bibendum tristique, erat arcu auctor eros, eu accumsan neque dolor sit amet quam. Donec scelerisque turpis nunc, at vulputate lorem dignissim sed. Aenean gravida, ligula a sodales laoreet, ante quam malesuada diam, ut feugiat quam velit nec sapien. Donec sem erat, auctor non eros ut, porta dignissim sapien. Integer felis lectus, molestie ut dui et, iaculis sollicitudin turpis. Sed auctor.`

type UI struct {
	app *tview.Application

	conversations       map[string]string
	currentConversation string

	table    *tview.Table
	textView *tview.TextArea
	input    *tview.TextArea
	flex     *tview.Flex
}

func NewUI() *UI {
	ui := &UI{
		app:                 tview.NewApplication(),
		conversations:       make(map[string]string),
		currentConversation: "New Conversation",
	}

	ui.conversations["New Conversation"] = ""
	ui.conversations["Conversation 1"] = ""
	ui.conversations["Conversation 2"] = ""

	ui.setupTable()
	ui.setupTextView()
	ui.setupInputField()
	ui.setupLayout()
	ui.setupEventHandlers()

	return ui
}

func (c *UI) setupTable() {
	c.table = tview.NewTable()
	c.table.SetBorders(false)
	c.table.SetSelectable(true, false)
	c.table.SetBackgroundColor(tcell.ColorDefault)

	row := 0
	for conv := range c.conversations {
		cell := tview.NewTableCell(conv).
			SetSelectable(true).
			SetTextColor(tcell.ColorWhite).
			SetBackgroundColor(tcell.ColorDefault)
		c.table.SetCell(row, 0, cell)
		row++
	}

	c.table.Select(0, 0)
}

func (c *UI) setupTextView() {
	c.textView = tview.NewTextArea()
	c.textView.SetBackgroundColor(tcell.ColorDefault)
	c.textView.SetBorder(false)
	c.updateTextView()
}

func (c *UI) setupInputField() {
	// TODO(nullswan): Make cursor blink
	c.input = tview.NewTextArea()
	c.input.SetBackgroundColor(tcell.ColorDefault)
	c.input.SetPlaceholder("Type your message here...")
	c.input.SetPlaceholderStyle(tcell.StyleDefault.Foreground(tcell.ColorGray))
	c.input.SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite))
	c.input.SetBorder(false)

	c.input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			modifiers := event.Modifiers()
			if modifiers&tcell.ModShift != 0 || modifiers&tcell.ModCtrl != 0 ||
				modifiers&tcell.ModMeta != 0 {
				// Insert a newline
				c.input.SetText(c.input.GetText()+"\n", true)
				return nil
			} else {
				// Submit the message
				inputText := c.input.GetText()
				if inputText != "" {
					if c.conversations[c.currentConversation] == "" {
						c.conversations[c.currentConversation] = inputText
					} else {
						c.conversations[c.currentConversation] += "\n" + inputText
					}
					c.updateTextView()
					c.input.SetText("", true)
				}
				return nil
			}
		}
		return event
	})
}

func (c *UI) setupLayout() {
	// Create a horizontal divider for the left panel
	dividerHorizontalLeft := tview.NewBox()
	dividerHorizontalLeft.SetDrawFunc(
		func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
			style := tcell.StyleDefault.
				Foreground(tcell.ColorGrey).
				Background(tcell.ColorDefault).
				Bold(true)

			// Draw a thick horizontal line
			for i := x; i < x+width; i++ {
				screen.SetContent(i, y, tcell.RuneHLine, nil, style)
				screen.SetContent(i, y+1, tcell.RuneHLine, nil, style)
			}
			return x, y, width, height
		},
	)

	// Create a text box below the conversation list
	conversationTextBox := tview.NewTextView()
	conversationTextBox.SetText("{Provider}\n{Delay}")
	conversationTextBox.SetBackgroundColor(tcell.ColorDefault)

	// Update the left panel layout
	leftFlex := tview.NewFlex()
	leftFlex.SetDirection(tview.FlexRow)
	leftFlex.AddItem(c.table, 0, 1, true)
	leftFlex.AddItem(dividerHorizontalLeft, 1, 0, false)
	leftFlex.AddItem(conversationTextBox, 3, 0, false)

	// Create a horizontal divider
	dividerHorizontal := tview.NewBox()
	dividerHorizontal.SetDrawFunc(
		func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
			style := tcell.StyleDefault.
				Foreground(tcell.ColorGrey).
				Background(tcell.ColorDefault).
				Bold(true)

			// Draw a thick horizontal line
			for i := x; i < x+width; i++ {
				screen.SetContent(i, y, tcell.RuneHLine, nil, style)
				screen.SetContent(i, y+1, tcell.RuneHLine, nil, style)
			}
			return x, y, width, height
		},
	)

	flexMain := tview.NewFlex()
	flexMain.SetDirection(tview.FlexRow)
	flexMain.AddItem(c.textView, 0, 1, false)
	flexMain.AddItem(dividerHorizontal, 1, 0, false)
	flexMain.AddItem(c.input, 5, 0, true)
	flexMain.SetBackgroundColor(tcell.ColorDefault)

	// Create a vertical divider
	divider := tview.NewBox()
	divider.SetDrawFunc(
		func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
			style := tcell.StyleDefault.
				Foreground(tcell.ColorGrey).
				Background(tcell.ColorDefault).
				Bold(true)

			// Draw a thick vertical line
			for i := y; i < y+height; i++ {
				screen.SetContent(x, i, tcell.RuneVLine, nil, style)
				screen.SetContent(x+1, i, tcell.RuneVLine, nil, style)
			}
			return x, y, width, height
		},
	)

	c.flex = tview.NewFlex()
	c.flex.AddItem(leftFlex, 20, 1, true)
	c.flex.AddItem(divider, 1, 0, false)
	c.flex.AddItem(flexMain, 0, 5, false)
	c.flex.SetBackgroundColor(tcell.ColorDefault)
}

func (c *UI) setupEventHandlers() {
	tableSelectedFunc := func(row, column int) {
		cell := c.table.GetCell(row, column)
		if cell == nil {
			return
		}
		selected := cell.Text
		if selected != "" {
			c.currentConversation = selected
			c.updateTextView()
			c.app.SetFocus(c.input)
		}
	}
	c.table.SetSelectedFunc(tableSelectedFunc)

	c.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			row, column := c.table.GetSelection()
			tableSelectedFunc(row, column)
			return nil
		}
		return event
	})

	c.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			switch c.app.GetFocus() {
			case c.input:
				c.app.SetFocus(c.textView)
			case c.textView:
				c.app.SetFocus(c.table)
			case c.table:
				c.app.SetFocus(c.input)
			}
			return nil
		}
		return event
	})

	// c.textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
	// 	switch event.Key() {
	// 	case tcell.KeyUp:
	// 		c.textView.ScrollUp()
	// 		return nil
	// 	case tcell.KeyDown:
	// 		c.textView.ScrollDown()
	// 		return nil
	// 	case tcell.KeyPgUp:
	// 		c.textView.ScrollPageUp()
	// 		return nil
	// 	case tcell.KeyPgDn:
	// 		c.textView.ScrollPageDown()
	// 		return nil
	// 	}
	// 	return event
	// })
}

func (c *UI) updateTextView() {
	content := c.conversations[c.currentConversation]
	if content == "" {
		content = "## " + c.currentConversation + "\n\n_Start your conversation..._"
	}

	content += longText + longText // Duplicate to make it longer
	c.textView.SetText(content, true)
}

func main() {
	ui := NewUI()

	ui.app.SetRoot(ui.flex, true)
	ui.app.EnablePaste(true)
	ui.app.EnableMouse(true)

	// Start with input field focused
	ui.app.SetFocus(ui.input)

	if err := ui.app.Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
