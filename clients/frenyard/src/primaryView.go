package src

import (
	"fmt"
	"strings"

	"github.com/uwu/frenyard/design"
	"github.com/uwu/frenyard/framework"
	"github.com/uwu/rethink/clients/rethinkgo"
)

func (app *UpApplication) ShowPrimaryView(thoughts []rethinkgo.Thought, warns ...string) {
	var warnings []string
	if len(warns) > 0 {
		warnings = warns
	}

	var slots []framework.FlexboxSlot

	fmt.Println(warnings)
	for _, warning := range warnings {
		slots = append(slots, framework.FlexboxSlot{
			Element: design.InformationPanel(design.InformationPanelDetails{
				Text: warning,
			}),
		})
	}

	thought := ""
	slots = append(slots, []framework.FlexboxSlot{
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
										// Ensure the teleportation affinity isn't set to any particular direction...
										app.GSInstant()

										// PUT whatever's in the textbox right into rethink.
										err := rethinkgo.PutThought(thought, app.Config.Name, app.Config.UploadKey)
										if err != nil {
											error := err.Error()
											fmt.Printf("Something went wrong while submitting a thought: %s", error)
											if strings.Contains(error, "401") {
												warnings = append(warnings, "You are not authorized to submit here.\nAre you on the right user, and is your upload key correct?")
											}
										}

										newThoughts, err := rethinkgo.GetThoughts(app.Config.Name)
										if err != nil {
											// This shouldn't ever happen...
											fmt.Printf("Something went wrong while getting the new thoughts: %s", err.Error())
											warnings = append(warnings, "Something went wrong while re-fetching your thoughts.\nThis should never happen.")
										}

										// If there are any warnings, display them. Otherwise, show the new thoughts.
										if warnings != nil {
											app.ShowPrimaryView(thoughts, warnings...)
										} else {
											app.ShowPrimaryView(newThoughts)
										}
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
	}...)

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
