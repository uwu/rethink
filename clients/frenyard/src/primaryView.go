package src

import (
	"github.com/uwu/frenyard/design"
	"github.com/uwu/frenyard/framework"
	"github.com/uwu/rethink/clients/frenyard/middle"
)

func (app *UpApplication) ShowPrimaryView() {
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
			}),
		},
		{
			Grow: 1,
		},
	}

	app.Teleport(design.LayoutDocument(design.Header{
		Title: "rethink",
		// Back: func() {
		// 	app.CachedPrimaryView = nil
		// 	app.GSLeftwards()
		// },
		BackIcon:    design.BackIconID,
		ForwardIcon: design.MenuIconID,
	}, framework.NewUIFlexboxContainerPtr(framework.FlexboxContainer{
		DirVertical: true,
		Slots:       slots,
	}), true))
}
