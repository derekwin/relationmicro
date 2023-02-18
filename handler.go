package main

import (
	"context"
	"fmt"
	"log"

	relation "github.com/ClubWeGo/relationmicro/kitex_gen/relation"
	"github.com/ClubWeGo/relationmicro/service"
)

// 响应码
const (
	// 服务器异常
	ERROR = 0
	// 正常响应
	SUCCESS = 1
	// 参数校验不通过
	VERIFY = 2
)

// 关注操作类型
const (
	// 关注
	FOLLOW = 1
	// 取关
	UNFOLLOW = 2
)

// CombineServiceImpl implements the last service interface defined in the IDL.
type CombineServiceImpl struct{}

// GetFollowListReqMethod implements the RelationServiceImpl interface.
func (s *CombineServiceImpl) GetFollowListReqMethod(ctx context.Context, request *relation.GetFollowListReq) (resp *relation.GetFollowListResp, err error) {
	myId := request.MyId
	targetId := request.TargetId
	// 参数校验
	if verifyMsg := VerifyFindFollowParam(myId, targetId); verifyMsg != nil {
		return &relation.GetFollowListResp{
			StatusCode: VERIFY,
			UserList:   []*relation.User{},
			Msg:        verifyMsg,
		}, nil
	}
	// myId 为空 isFollow全为false 无影响
	followList, err := service.FindFollowList(*myId, targetId)

	if err != nil {
		return &relation.GetFollowListResp{
			StatusCode: ERROR,
			UserList:   []*relation.User{},
		}, nil
	}

	// 封装响应
	respUserList := make([]*relation.User, len(followList))
	for i, followUser := range followList {
		fmt.Println(followUser)
		respUserList[i] = &relation.User{
			Id:            followUser.Id,
			Name:          followUser.Name,
			FollowCount:   followUser.FollowCount,
			FollowerCount: followUser.FollowerCount,
			IsFollow:      followUser.IsFollow,
		}
	}
	return &relation.GetFollowListResp{
		StatusCode: SUCCESS,
		UserList:   respUserList,
	}, err

}

// GetFollowerListMethod implements the RelationServiceImpl interface.
func (s *CombineServiceImpl) GetFollowerListMethod(ctx context.Context, request *relation.GetFollowerListReq) (resp *relation.GetFollowerListResp, err error) {
	myId := request.MyId
	targetId := request.TargetId
	// 参数校验
	if verifyMsg := VerifyFindFollowParam(myId, targetId); verifyMsg != nil {
		return &relation.GetFollowerListResp{
			StatusCode: VERIFY,
			UserList:   []*relation.User{},
			Msg:        verifyMsg,
		}, nil
	}
	// myId 为空 isFollow全为false 无影响
	followerList, err := service.FindFollowerList(*myId, targetId)
	if err != nil {
		return &relation.GetFollowerListResp{
			StatusCode: ERROR,
			UserList:   []*relation.User{},
		}, err
	}
	// 封装响应
	respUserList := make([]*relation.User, len(followerList))
	for i, followerUser := range followerList {
		respUserList[i] = &relation.User{
			Id:            followerUser.Id,
			Name:          followerUser.Name,
			FollowCount:   followerUser.FollowCount,
			FollowerCount: followerUser.FollowerCount,
			IsFollow:      followerUser.IsFollow,
		}
	}
	return &relation.GetFollowerListResp{
		StatusCode: SUCCESS,
		UserList:   respUserList,
	}, nil
}

// GetAllMessageMethod implements the MessageServiceImpl interface.
func (s *CombineServiceImpl) GetAllMessageMethod(ctx context.Context, request *relation.GetAllMessageReq) (resp *relation.GetAllMessageResp, err error) {
	// TODO: Your code here...
	// service层拿数据
	msgs, err := service.GetAllP2PMsg(request.UserId, request.ToUserId)

	if err != nil {
		return &relation.GetAllMessageResp{
			Status: false,
			Msg:    []*relation.Message{}, //返回空消息
		}, nil
	}

	respMsg := make([]*relation.Message, len(msgs))
	for index, msg := range msgs {
		createTimeString := msg.Create_at.Format("2006-01-02")
		respMsg[index] = &relation.Message{
			Id:         msg.Id,
			FromUserId: msg.UserId,
			ToUserId:   msg.ToUserId,
			Content:    msg.Content,
			CreateTime: &createTimeString,
		}
	}
	return &relation.GetAllMessageResp{
		Status: true,
		Msg:    respMsg,
	}, nil
}

// SendMessageMethod implements the MessageServiceImpl interface.
func (s *CombineServiceImpl) SendMessageMethod(ctx context.Context, request *relation.SendMessageReq) (resp *relation.SendMessageResp, err error) {
	// TODO: Your code here...
	// service层拿数据
	_, err = service.SendP2PMsg(request.UserId, request.ToUserId, request.Content)
	if err != nil {
		return &relation.SendMessageResp{
			Status: false,
		}, err
	}
	return &relation.SendMessageResp{
		Status: true,
	}, nil
}

// FollowMethod implements the RelationServiceImpl interface.
func (s *CombineServiceImpl) FollowMethod(ctx context.Context, request *relation.FollowReq) (resp *relation.FollowResp, err error) {
	// 校验请求参数
	if verifyMsg := VerifyFollowParam(request.MyUid, request.TargetUid); verifyMsg != nil {
		resp = &relation.FollowResp{
			StatusCode: VERIFY,
			Msg:        verifyMsg,
		}
	}
	// 关注类型
	actionType := request.ActionType
	resp = &relation.FollowResp{
		StatusCode: SUCCESS,
	}
	var errMsg error

	if actionType == FOLLOW {
		// 关注
		errMsg = service.Follow(request.MyUid, request.TargetUid)
	} else if actionType == UNFOLLOW {
		// 取关
		errMsg = service.UnFollow(request.MyUid, request.TargetUid)
	} else {
		// 其他情况算正常操作 反正对数据无影响，或者上游直接禁掉
		return resp, nil
	}
	if errMsg != nil {
		resp.StatusCode = ERROR
		log.Printf("FollowMethod err：%s", err)
	}
	return resp, nil
}

// GetFollowInfoMethod implements the RelationServiceImpl interface.
func (s *CombineServiceImpl) GetFollowInfoMethod(ctx context.Context, req *relation.GetFollowInfoReq) (resp *relation.GetFollowInfoResp, err error) {
	myUid := req.MyUid
	targetUid := req.TargetUid

	// 校验请求参数
	if verifyMsg := VerifyFindFollowParam(myUid, targetUid); verifyMsg != nil {
		return &relation.GetFollowInfoResp{
			StatusCode: VERIFY,
			Msg:        verifyMsg,
		}, nil
	}

	// 如果请求端没有携带用户信息 默认返回未关注
	isFollow := false
	if myUid != nil {
		isFollow = service.FindIsFollow(*myUid, targetUid)
	}
	return &relation.GetFollowInfoResp{
		StatusCode: SUCCESS,
		FollowInfo: &relation.FollowInfo{
			FollowCount:   service.FindFollowCount(targetUid),
			FollowerCount: service.FindFollowerCount(targetUid),
			IsFollow:      isFollow,
		},
	}, nil
}

// 校验关注的非法请求参数
func VerifyFollowParam(myUid int64, targetUid int64) *string {
	var errMsg *string = nil
	if myUid == targetUid {
		*errMsg = "two uid the same, not allow!"
	}
	return errMsg
}

// 校验查询关注信息的非法请求参数
func VerifyFindFollowParam(myUid *int64, targetUid int64) *string {
	return nil
}

// GetFriendListMethod implements the RelationServiceImpl interface.
func (s *CombineServiceImpl) GetFriendListMethod(ctx context.Context, request *relation.GetFriendListReq) (resp *relation.GetFriendListResp, err error) {
	// TODO: Your code here...
	return
}
