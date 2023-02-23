package src

import (
	"github.com/uwu/frenyard/design"
	"github.com/uwu/frenyard/framework"
)

func (app *UpApplication) ShowPrimaryView() {
	test1 := ""
	test2 := ""
	slots := []framework.FlexboxSlot{
		{
			Element: design.NewUITextboxPtr("Name", &test1),
		},
		{
			Grow: 1,
		},
		{
			Element: framework.NewUIFlexboxContainerPtr(framework.FlexboxContainer{
				DirVertical: false,
				Slots: []framework.FlexboxSlot{
					{
						Grow: 1,
					},
					{
						Element: design.NewUITextboxPtr("Name", &test1),
						// Shrink:  1,
					},
					// {
					// 	Basis:  frenyard.Scale(design.DesignScale, 32),
					// 	Shrink: 1,
					// },
					{
						Element: design.NewUITextboxPtr("Name", &test2),
						// Shrink:  1,
					},
					{
						Grow: 1,
					},
				},
			}),
		},
		{
			Grow: 1,
		},
	}

	app.Teleport(design.LayoutDocument(design.Header{
		Title: "rethink",
		Back: func() {
			app.CachedPrimaryView = nil
			app.GSLeftwards()
		},
		BackIcon:    design.BackIconID,
		ForwardIcon: design.MenuIconID,
	}, framework.NewUIFlexboxContainerPtr(framework.FlexboxContainer{
		DirVertical: true,
		Slots:       slots,
	}), true))
}
