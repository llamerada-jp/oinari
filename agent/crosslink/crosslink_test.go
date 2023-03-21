package crosslink

import (
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

	mpxRoot.SetHandler("goFunc", NewFuncHandler(func(data string, tags map[string]string, writer ResponseWriter) {
		g.Expect(tags[TAG_PATH]).Should(Equal("goFunc"))
		g.Expect(tags[TAG_LEAF]).Should(Equal(""))

		crosslink.Call("jsFunc", []byte("request js"), map[string]string{
			"type": "success",
		}, func(result []byte, err error) {
			g.Expect(result).Should(Equal("result js success"))
			g.Expect(err).ShouldNot(HaveOccurred())

			crosslink.Call("jsFunc", []byte("request js"), map[string]string{
				"type": "failure",
			}, func(result []byte, err error) {
				g.Expect(result).Should(BeEmpty())
				g.Expect(err).Should(HaveOccurred())

				writer.ReplySuccess("result go func1")
			})
		})
	}))

	// get result async
	resultChan := make(chan bool)
	js.Global().Get("crosslinkGoTest").Set("finToGo", js.FuncOf(func(this js.Value, args []js.Value) any {
		resultChan <- args[0].Bool()
		return nil
	}))

	// run js test module
	go js.Global().Get("crosslinkGoTest").Call("runByGo")

	// wait result
	ti := time.NewTicker(time.Second)
	defer ti.Stop()
	for {
		select {
		case <-ti.C:
			// wait loop
		case result := <-resultChan:
			g.Expect(result).Should(BeTrue())
			return
		}
	}
}
