package src

import (
	"fmt"
	"os"
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
		thoughts, err := rethinkgo.GetThoughts(app.Config.Name)
		if err != nil {
			error := err.Error()
			fmt.Printf("Something went wrong while fetching thoughts: %s\n", error)
			if strings.HasSuffix(error, ": no such host") {
				warnings = append(warnings, fmt.Sprintf("The provided API endpoint does not exist or is invalid. (\"%s\")", os.Getenv("RETHINK_API")))
			}
			if strings.Contains(error, "unsupported protocol scheme") {
				warnings = append(warnings, fmt.Sprintf("The provided API endpoint is invalid. (\"%s\")", os.Getenv("RETHINK_API")))
			}
			if strings.HasSuffix(error, ": connection refused") {
				warnings = append(warnings, "Rethink can't be reached.")
			}
			if strings.Contains(error, "404") {
				warnings = append(warnings, "This user couldn't be found.")
			}
		}
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
		Title: "welcome | rethink",
	}, framework.NewUIFlexboxContainerPtr(framework.FlexboxContainer{
		DirVertical: true,
		Slots:       slots,
	}), true))
}
