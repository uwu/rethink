package src

import (
	"fmt"
	"github.com/uwu/frenyard/design"
	"github.com/uwu/frenyard/framework"
	"github.com/uwu/rethink/clients/rethinkgo"
)

func (app *UpApplication) ShowPrimaryView(thoughts []rethinkgo.Thought) {
	//thoughts, err := rethinkgo.GetThoughts("fucker")
	//if err != nil {
	//	fmt.Printf("Failed fetching thoughts: %s", err.Error())
	//}
	fmt.Println(thoughts)

	app.Teleport(design.LayoutDocument(design.Header{
		Title: "rethink | welcome",
	}, framework.NewUIFlexboxContainerPtr(framework.FlexboxContainer{
		DirVertical: true,
		Slots:       []framework.FlexboxSlot{},
	}), true))
}
