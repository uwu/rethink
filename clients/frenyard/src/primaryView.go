package src

import (
	"fmt"
	"github.com/uwu/frenyard/design"
	"github.com/uwu/frenyard/framework"
	"github.com/uwu/rethink/clients/rethinkgo"
)

func (app *UpApplication) ShowPrimaryView(thoughts []rethinkgo.Thought) {
	var slots []framework.FlexboxSlot

	for _, thought := range thoughts {
		slots = append(slots, framework.FlexboxSlot{
			Element: design.ListItem(design.ListItemDetails{
				Text:    thought.Content,
				Subtext: thought.Date.String(),
			}),
		})
	}

	app.Teleport(design.LayoutDocument(design.Header{
		Title: fmt.Sprintf("%s | rethink", app.Config.Name),
	}, framework.NewUIFlexboxContainerPtr(framework.FlexboxContainer{
		DirVertical: true,
		Slots:       slots,
	}), true))
}
