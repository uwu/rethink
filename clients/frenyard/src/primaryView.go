package src

import (
	"fmt"
	"github.com/uwu/frenyard/design"
	"github.com/uwu/frenyard/framework"
	"github.com/uwu/rethink/clients/rethinkgo"
)

func (app *UpApplication) ShowPrimaryView(thoughts []rethinkgo.Thought) {
	var slots []framework.FlexboxSlot
	thought := ""
	slots := []framework.FlexboxSlot{
		{
			Element: framework.NewUIFlexboxContainerPtr(framework.FlexboxContainer{
				DirVertical: true,
				Slots: []framework.FlexboxSlot{
					{
						Element: design.NewUITextareaPtr("Think of something...", &thought),
					},
					{
						Basis: 25,
					},
					{
						Element: framework.NewUIFlexboxContainerPtr(framework.FlexboxContainer{
							DirVertical: false,
							Slots: []framework.FlexboxSlot{
								{
									Grow: 1,
								},
								{
									Element: design.ButtonAction(design.ThemeOkActionButton, "Submit", func() {
										err := rethinkgo.PutThought(thought, app.Config.Name, app.Config.UploadKey)
										app.GSInstant()
										newThoughts, err := rethinkgo.GetThoughts(app.Config.Name)
										if err != nil {
											fmt.Println("Something went wrong")
										}
										app.ShowPrimaryView(newThoughts)
									}),
								},
							},
						}),
					},
				},
			}),
		},
		{
			Basis: 30,
		},
	}

	for i, thought := range thoughts {
		slots = append(slots, framework.FlexboxSlot{
			Element: design.ListItem(design.ListItemDetails{
				Text:    thought.Content,
				Subtext: thought.Date.String(),
			}),
		})

		if i < len(thoughts) {
			slots = append(slots, framework.FlexboxSlot{
				Basis: 45,
			})
		}
	}

	app.Teleport(design.LayoutDocument(design.Header{
		Title: fmt.Sprintf("%s | rethink", app.Config.Name),
		Back: func() {
			app.CachedPrimaryView = nil
			app.CachedThoughts = nil
			app.GSLeftwards()
			app.ShowLoginForm()
		},
		BackIcon: design.BackIconID,
	}, framework.NewUIFlexboxContainerPtr(framework.FlexboxContainer{
		DirVertical: true,
		Slots:       slots,
	}), true))
}
