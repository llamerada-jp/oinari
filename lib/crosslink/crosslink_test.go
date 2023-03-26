package crosslink

import (
	"encoding/json"
	"syscall/js"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestCrosslink(t *testing.T) {
	g := NewGomegaWithT(t)

	// setup
	mpxRoot := NewMultiPlexer()
	crosslink := NewCrosslink("crosslinkGo", mpxRoot)

	mpxRoot.SetHandler("goFunc", NewFuncHandler(func(data *string, tags map[string]string, writer ResponseWriter) {
		g.Expect(tags[TAG_PATH]).Should(Equal("goFunc"))
		g.Expect(tags[TAG_LEAF]).Should(Equal(""))

		crosslink.Call("jsFunc", "request js1", map[string]string{
			"type": "success",
		}, func(responseRaw []byte, err error) {
			g.Expect(err).ShouldNot(HaveOccurred())

			var response string
			err = json.Unmarshal(responseRaw, &response)
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(response).Should(Equal("response js success"))

			crosslink.Call("jsFunc", "request js2", map[string]string{
				"type": "failure",
			}, func(responseRaw []byte, err error) {
				g.Expect(responseRaw).Should(BeEmpty())
				g.Expect(err).Should(HaveOccurred())

				writer.ReplySuccess("response go func1")
			})
		})
	}))

	// get response async
	responseChan := make(chan bool)
	js.Global().Get("crosslinkGoTest").Set("finToGo", js.FuncOf(func(this js.Value, args []js.Value) any {
		responseChan <- args[0].Bool()
		return nil
	}))

	// run js test module
	go js.Global().Get("crosslinkGoTest").Call("runByGo")

	// wait response
	ti := time.NewTicker(time.Second)
	defer ti.Stop()
	for {
		select {
		case <-ti.C:
			// wait loop
		case response := <-responseChan:
			g.Expect(response).Should(BeTrue())
			return
		}
	}
}
