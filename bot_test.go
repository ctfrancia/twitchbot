package bot

import (
	"testing"
)

var bb BasicBot

func TestHandleChatPrivMsg(t *testing.T) {
	handleChatPrivMsg([]string{"cheer100", "hello"}, &bb)
	// if x != 5 {
	// t.Error("Expected, 5, got: ", x)
	// }
}
