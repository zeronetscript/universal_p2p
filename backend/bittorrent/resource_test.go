package bittorrent

import (
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

const lastAccessFile = "testing/lastAccess.json"

func TestSaveLoad(t *testing.T) {

	toSave := make(map[string]*Resource)

	infoHash := "d119f5f2635ae774cde6b60ae67e1dfa9016811f"

	now := time.Now()

	var to torrent.Torrent

	sub := &Resource{
		lastAccess: now,
		SubFile:    &torrent.File{},
	}

	tmp := make(map[string]*Resource)

	tmp["mp4.mp4"] = sub

	res := Resource{
		Torrent:      &to,
		lastAccess:   now,
		SubResources: &tmp,
	}

	sub.rootRes = &res

	toSave[infoHash] = &res

	fmt.Print("save")
	er := saveLastAccessFile(&toSave, lastAccessFile)
	fmt.Print("save ok")

	assert.Empty(t, er)

	fmt.Print("load")
	loaded := make(map[string]*Resource)
	err := loadLastAccessFile(&loaded, lastAccessFile)
	fmt.Print("loado k")

	assert.Empty(t, err)

	v, ok := (loaded)[infoHash]

	assert.Equal(t, ok, true, "should load")

	assert.NotEmpty(t, v.SubResources)

	assert.Equal(t, len(*v.SubResources), 1)

}

func TestMain(m *testing.M) {
	os.MkdirAll("testing", os.ModeDir|os.ModePerm)
	code := m.Run()
	os.Exit(code)
}
