// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: like.proto

package newspb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type LikeNewsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	NewsId        string                 `protobuf:"bytes,1,opt,name=news_id,json=newsId,proto3" json:"news_id,omitempty"`
	UserId        string                 `protobuf:"bytes,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *LikeNewsRequest) Reset() {
	*x = LikeNewsRequest{}
	mi := &file_like_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *LikeNewsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LikeNewsRequest) ProtoMessage() {}

func (x *LikeNewsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_like_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LikeNewsRequest.ProtoReflect.Descriptor instead.
func (*LikeNewsRequest) Descriptor() ([]byte, []int) {
	return file_like_proto_rawDescGZIP(), []int{0}
}

func (x *LikeNewsRequest) GetNewsId() string {
	if x != nil {
		return x.NewsId
	}
	return ""
}

func (x *LikeNewsRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

type LikeNewsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	LikeCount     int64                  `protobuf:"varint,2,opt,name=like_count,json=likeCount,proto3" json:"like_count,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *LikeNewsResponse) Reset() {
	*x = LikeNewsResponse{}
	mi := &file_like_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *LikeNewsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LikeNewsResponse) ProtoMessage() {}

func (x *LikeNewsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_like_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LikeNewsResponse.ProtoReflect.Descriptor instead.
func (*LikeNewsResponse) Descriptor() ([]byte, []int) {
	return file_like_proto_rawDescGZIP(), []int{1}
}

func (x *LikeNewsResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *LikeNewsResponse) GetLikeCount() int64 {
	if x != nil {
		return x.LikeCount
	}
	return 0
}

type UnlikeNewsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	NewsId        string                 `protobuf:"bytes,1,opt,name=news_id,json=newsId,proto3" json:"news_id,omitempty"`
	UserId        string                 `protobuf:"bytes,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UnlikeNewsRequest) Reset() {
	*x = UnlikeNewsRequest{}
	mi := &file_like_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UnlikeNewsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UnlikeNewsRequest) ProtoMessage() {}

func (x *UnlikeNewsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_like_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UnlikeNewsRequest.ProtoReflect.Descriptor instead.
func (*UnlikeNewsRequest) Descriptor() ([]byte, []int) {
	return file_like_proto_rawDescGZIP(), []int{2}
}

func (x *UnlikeNewsRequest) GetNewsId() string {
	if x != nil {
		return x.NewsId
	}
	return ""
}

func (x *UnlikeNewsRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

type UnlikeNewsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	LikeCount     int64                  `protobuf:"varint,2,opt,name=like_count,json=likeCount,proto3" json:"like_count,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UnlikeNewsResponse) Reset() {
	*x = UnlikeNewsResponse{}
	mi := &file_like_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UnlikeNewsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UnlikeNewsResponse) ProtoMessage() {}

func (x *UnlikeNewsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_like_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UnlikeNewsResponse.ProtoReflect.Descriptor instead.
func (*UnlikeNewsResponse) Descriptor() ([]byte, []int) {
	return file_like_proto_rawDescGZIP(), []int{3}
}

func (x *UnlikeNewsResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *UnlikeNewsResponse) GetLikeCount() int64 {
	if x != nil {
		return x.LikeCount
	}
	return 0
}

type GetLikesCountRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	NewsId        string                 `protobuf:"bytes,1,opt,name=news_id,json=newsId,proto3" json:"news_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetLikesCountRequest) Reset() {
	*x = GetLikesCountRequest{}
	mi := &file_like_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetLikesCountRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetLikesCountRequest) ProtoMessage() {}

func (x *GetLikesCountRequest) ProtoReflect() protoreflect.Message {
	mi := &file_like_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetLikesCountRequest.ProtoReflect.Descriptor instead.
func (*GetLikesCountRequest) Descriptor() ([]byte, []int) {
	return file_like_proto_rawDescGZIP(), []int{4}
}

func (x *GetLikesCountRequest) GetNewsId() string {
	if x != nil {
		return x.NewsId
	}
	return ""
}

type GetLikesCountResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	LikeCount     int64                  `protobuf:"varint,1,opt,name=like_count,json=likeCount,proto3" json:"like_count,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetLikesCountResponse) Reset() {
	*x = GetLikesCountResponse{}
	mi := &file_like_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetLikesCountResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetLikesCountResponse) ProtoMessage() {}

func (x *GetLikesCountResponse) ProtoReflect() protoreflect.Message {
	mi := &file_like_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetLikesCountResponse.ProtoReflect.Descriptor instead.
func (*GetLikesCountResponse) Descriptor() ([]byte, []int) {
	return file_like_proto_rawDescGZIP(), []int{5}
}

func (x *GetLikesCountResponse) GetLikeCount() int64 {
	if x != nil {
		return x.LikeCount
	}
	return 0
}

var File_like_proto protoreflect.FileDescriptor

const file_like_proto_rawDesc = "" +
	"\n" +
	"\n" +
	"like.proto\x12\x04news\"C\n" +
	"\x0fLikeNewsRequest\x12\x17\n" +
	"\anews_id\x18\x01 \x01(\tR\x06newsId\x12\x17\n" +
	"\auser_id\x18\x02 \x01(\tR\x06userId\"K\n" +
	"\x10LikeNewsResponse\x12\x18\n" +
	"\asuccess\x18\x01 \x01(\bR\asuccess\x12\x1d\n" +
	"\n" +
	"like_count\x18\x02 \x01(\x03R\tlikeCount\"E\n" +
	"\x11UnlikeNewsRequest\x12\x17\n" +
	"\anews_id\x18\x01 \x01(\tR\x06newsId\x12\x17\n" +
	"\auser_id\x18\x02 \x01(\tR\x06userId\"M\n" +
	"\x12UnlikeNewsResponse\x12\x18\n" +
	"\asuccess\x18\x01 \x01(\bR\asuccess\x12\x1d\n" +
	"\n" +
	"like_count\x18\x02 \x01(\x03R\tlikeCount\"/\n" +
	"\x14GetLikesCountRequest\x12\x17\n" +
	"\anews_id\x18\x01 \x01(\tR\x06newsId\"6\n" +
	"\x15GetLikesCountResponse\x12\x1d\n" +
	"\n" +
	"like_count\x18\x01 \x01(\x03R\tlikeCountB@Z>github.com/Abdurahmanit/GroupProject/news-service/proto;newspbb\x06proto3"

var (
	file_like_proto_rawDescOnce sync.Once
	file_like_proto_rawDescData []byte
)

func file_like_proto_rawDescGZIP() []byte {
	file_like_proto_rawDescOnce.Do(func() {
		file_like_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_like_proto_rawDesc), len(file_like_proto_rawDesc)))
	})
	return file_like_proto_rawDescData
}

var file_like_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_like_proto_goTypes = []any{
	(*LikeNewsRequest)(nil),       // 0: news.LikeNewsRequest
	(*LikeNewsResponse)(nil),      // 1: news.LikeNewsResponse
	(*UnlikeNewsRequest)(nil),     // 2: news.UnlikeNewsRequest
	(*UnlikeNewsResponse)(nil),    // 3: news.UnlikeNewsResponse
	(*GetLikesCountRequest)(nil),  // 4: news.GetLikesCountRequest
	(*GetLikesCountResponse)(nil), // 5: news.GetLikesCountResponse
}
var file_like_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_like_proto_init() }
func file_like_proto_init() {
	if File_like_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_like_proto_rawDesc), len(file_like_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_like_proto_goTypes,
		DependencyIndexes: file_like_proto_depIdxs,
		MessageInfos:      file_like_proto_msgTypes,
	}.Build()
	File_like_proto = out.File
	file_like_proto_goTypes = nil
	file_like_proto_depIdxs = nil
}
