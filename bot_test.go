package bot

import (
	"testing"
)

var bb BasicBot

func TestHandleChatPrivMsg(t *testing.T) {
	handleChatPrivMsg([]string{"cheer100", "hello", "test", "third"}, &bb)
}
