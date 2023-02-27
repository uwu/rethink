package src

import (
	"fmt"
	"strings"

	"github.com/uwu/frenyard/design"
	"github.com/uwu/frenyard/framework"
	"github.com/uwu/rethink/clients/frenyard/middle"
	"github.com/uwu/rethink/clients/rethinkgo"
)

func (app *UpApplication) ShowPreface() {
	var warnings []string

	app.ShowWaiter("Loading...", func(progress func(string)) {
		progress("Fetching thoughts...")
		fmt.Println(app.Config)
		thoughts, err := rethinkgo.GetThoughts(app.Config.Name)
		if err != nil {
			fmt.Printf("Something went wrong while fetching thoughts: %s\n", err.Error())
			if strings.HasSuffix(err.Error(), ": connection refused") {
				warnings = append(warnings, "Rethink can't be reached.")
			}
			if strings.ContainsAny(err.Error(), "404") {
				warnings = append(warnings, "This user couldn't be found.")
			}
		}
		fmt.Println(thoughts)
		app.CachedThoughts = thoughts
	}, func() {
		if app.CachedThoughts == nil {
			app.ShowLoginForm(warnings...)
		} else {
			app.CachedPrimaryView = nil
			app.GSRightwards()
			app.ShowPrimaryView(app.CachedThoughts)
		}
	})
}

func (app *UpApplication) ShowLoginForm(warns ...string) {
	var warnings []string
	if len(warns) > 0 {
		warnings = warns
	}

	name := ""
	uploadKey := ""
	config := middle.ReadConfig()
	slots := []framework.FlexboxSlot{}

	for _, warning := range warnings {
		slots = append(slots, framework.FlexboxSlot{
			Element: design.InformationPanel(design.InformationPanelDetails{
				Text: warning,
			}),
		})
	}

	slots = append(slots, []framework.FlexboxSlot{
		{
			Grow: 1,
		},
		{
			Element: design.NewUITextboxPtr("Name", &name, config.Name),
		},
		{
			Basis: 25,
		},
		{
			Element: design.NewUITextboxPtr("Upload Key", &uploadKey, config.UploadKey),
		},
		{
			Basis: 25,
		},
		{
			Element: design.ButtonAction(design.ThemeOkActionButton, "Confirm", func() {
				app.Config.Name = name
				app.Config.UploadKey = uploadKey
				middle.WriteConfig(app.Config)
				app.GSInstant()
				app.ShowPreface()
			}),
		},
		{
			Grow: 1,
		},
	}...)

	app.Teleport(design.LayoutDocument(design.Header{
		Title: "rethink | welcome",
	}, framework.NewUIFlexboxContainerPtr(framework.FlexboxContainer{
		DirVertical: true,
		Slots:       slots,
	}), true))
}
