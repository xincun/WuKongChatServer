package label

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/WuKongIM/WuKongChatServer/internal/api/user"
	"github.com/WuKongIM/WuKongChatServer/internal/testutil"
	"github.com/WuKongIM/WuKongChatServer/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	f := New(ctx)
	f.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	memberUIDs := make([]string, 0)
	memberUIDs = append(memberUIDs, "111")
	memberUIDs = append(memberUIDs, "222")
	label := &labelReq{
		Name:       "标签1",
		MemberUIDs: memberUIDs,
	}
	req, _ := http.NewRequest("POST", "/v1/label", bytes.NewReader([]byte(util.ToJson(label))))
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDelete(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	l := New(ctx)
	l.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	label := &model{
		UID:        testutil.UID,
		Name:       "标签1",
		MemberUids: "111,222,333",
	}
	id, err := l.db.insert(label)
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/v1/label/%d", id), nil)
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestList(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	l := New(ctx)
	l.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	userservice := user.NewService(ctx)

	err = userservice.AddFriend(testutil.UID, &user.FriendReq{
		UID:   testutil.UID,
		ToUID: "111",
	})
	assert.NoError(t, err)
	err = userservice.AddFriend(testutil.UID, &user.FriendReq{
		UID:   testutil.UID,
		ToUID: "222",
	})
	assert.NoError(t, err)

	err = userservice.AddUser(&user.AddUserReq{
		UID:  "222",
		Name: "222",
	})
	assert.NoError(t, err)
	err = userservice.AddUser(&user.AddUserReq{
		UID:  "111",
		Name: "111",
	})
	assert.NoError(t, err)
	label := &model{
		UID:        testutil.UID,
		Name:       "标签1",
		MemberUids: "111,222",
	}
	_, err = l.db.insert(label)
	assert.NoError(t, err)
	req, _ := http.NewRequest("GET", "/v1/label", nil)
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDetail(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	l := New(ctx)
	l.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	userservice := user.NewService(ctx)

	err = userservice.AddFriend(testutil.UID, &user.FriendReq{
		UID:   testutil.UID,
		ToUID: "111",
	})
	assert.NoError(t, err)
	err = userservice.AddFriend(testutil.UID, &user.FriendReq{
		UID:   testutil.UID,
		ToUID: "222",
	})
	assert.NoError(t, err)

	err = userservice.AddUser(&user.AddUserReq{
		UID:  "222",
		Name: "222",
	})
	assert.NoError(t, err)
	err = userservice.AddUser(&user.AddUserReq{
		UID:  "111",
		Name: "111",
	})
	assert.NoError(t, err)
	label := &model{
		UID:        testutil.UID,
		Name:       "标签1",
		MemberUids: "111,222",
	}
	id, err := l.db.insert(label)
	assert.NoError(t, err)
	req, _ := http.NewRequest("GET", fmt.Sprintf("/v1/label/%d", id), nil)
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
