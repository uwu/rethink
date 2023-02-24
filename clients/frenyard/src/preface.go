package src

import (
	"fmt"
	"github.com/uwu/frenyard/design"
	"github.com/uwu/frenyard/framework"
	"github.com/uwu/rethink/clients/frenyard/middle"
	"github.com/uwu/rethink/clients/rethinkgo"
)

func (app *UpApplication) ShowPreface() {
	app.ShowWaiter("Loading...", func(progress func(string)) {
		progress("Fetching thoughts...")
		fmt.Println(app.Config)
		thoughts, err := rethinkgo.GetThoughts(app.Config.Name)
		if err != nil {
			fmt.Printf("Something went wrong while fetching thoughts: %s\n", err.Error())
		}
		fmt.Println(thoughts)
		app.CachedThoughts = thoughts
	}, func() {
		fmt.Println(app.CachedPrimaryView)
		if app.CachedThoughts == nil {
			app.ShowLoginForm()
		} else {
			app.CachedPrimaryView = nil
			app.GSRightwards()
			app.ShowPrimaryView(app.CachedThoughts)
		}
	})
}

func (app *UpApplication) ShowLoginForm() {
	name := ""
	uploadKey := ""
	slots := []framework.FlexboxSlot{
		{
			Grow: 1,
		},
		{
			Element: design.NewUITextboxPtr("Name", &name),
		},
		{
			Basis: 25,
		},
		{
			Element: design.NewUITextboxPtr("Upload Key", &uploadKey),
		},
		{
			Grow: 1,
		},
		{
			Element: design.ButtonAction(design.ThemeOkActionButton, "Confirm", func() {
				app.Config.Name = name
				app.Config.UploadKey = uploadKey
				middle.WriteConfig(app.Config)
				app.ShowPreface()
			}),
		},
		{
			Grow: 1,
		},
	}

	app.Teleport(design.LayoutDocument(design.Header{
		Title: "rethink | welcome",
	}, framework.NewUIFlexboxContainerPtr(framework.FlexboxContainer{
		DirVertical: true,
		Slots:       slots,
	}), true))
}
