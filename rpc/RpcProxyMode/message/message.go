package message

import (
	"bytes"
	"encoding/binary"
)

const (
	splitter     = '\n'
	pairSplitter = '\r'
)

type Request struct {
	HeadLength  uint32            `json:"head_length,omitempty"`
	BodyLength  uint32            `json:"body_length,omitempty"`
	MessageId   uint32            `json:"message_id,omitempty"`
	Version     byte              `json:"version,omitempty"`
	Compressor  byte              `json:"compressor,omitempty"`
	Serializer  byte              `json:"serializer,omitempty"`
	ServiceName string            `json:"service_name,omitempty"`
	MethodName  string            `json:"method_name,omitempty"`
	Meta        map[string]string `json:"meta,omitempty"`
	Data        []byte            `json:"data,omitempty"`
}

//计算头部长度
func (r *Request) CalHeadLength() {
	// 固定的 15 个字节
	res := 15
	res = res + len(r.ServiceName)
	// 我要加个分隔符 \n
	res += 1
	res = res + len(r.MethodName)
	res += 1

	// 加一个分隔符
	// meta
	for key, value := range r.Meta {
		// + 1 是为了区分 key 和 value
		res = res + len(key) + 1 + len(value) + 1
	}
	// |key|value|key|value

	r.HeadLength = uint32(res)
}

func EncodeReq(req *Request) []byte {
	// 分配内存
	bs := make([]byte, req.HeadLength+req.BodyLength)

	cur := bs
	// 头四个字节
	binary.BigEndian.PutUint32(cur[:4], req.HeadLength)
	cur = cur[4:]

	binary.BigEndian.PutUint32(cur[:4], req.BodyLength)
	cur = cur[4:]

	binary.BigEndian.PutUint32(cur[:4], req.MessageId)
	cur = cur[4:]

	cur[0] = req.Version
	cur = cur[1:]

	cur[0] = req.Compressor
	cur = cur[1:]

	cur[0] = req.Serializer
	cur = cur[1:]

	copy(cur, req.ServiceName)
	// 加个分隔符
	cur[len(req.ServiceName)] = splitter
	cur = cur[len(req.ServiceName)+1:]

	copy(cur, req.MethodName)
	// 加个分隔符
	cur[len(req.MethodName)] = splitter
	cur = cur[len(req.MethodName)+1:]

	for key, value := range req.Meta {
		copy(cur, key)
		// 加个分隔符
		cur[len(key)] = pairSplitter
		cur = cur[len(key)+1:]

		copy(cur, value)
		// 加个分隔符
		cur[len(value)] = splitter
		cur = cur[len(value)+1:]
	}

	copy(cur, req.Data)
	return bs
}
func DecodeReq(data []byte) *Request {
	req := &Request{}
	req.HeadLength = binary.BigEndian.Uint32(data[:4])
	req.BodyLength = binary.BigEndian.Uint32(data[4:8])
	req.MessageId = binary.BigEndian.Uint32(data[8:12])
	req.Version = data[12]
	req.Compressor = data[13]
	req.Serializer = data[14]

	// 是头部剩余数据
	head := data[15:req.HeadLength]
	index := bytes.IndexByte(head, splitter)
	req.ServiceName = string(head[:index])

	// 加1 是为了跳掉分隔符
	head = head[index+1:]
	index = bytes.IndexByte(head, splitter)
	req.MethodName = string(head[:index])

	// 加1 是为了跳掉分隔符
	head = head[index+1:]
	if len(head) > 0 {
		metaMap := make(map[string]string)
		index = bytes.IndexByte(head, splitter)
		// 切出来了
		for index != -1 {
			pair := head[:index]
			// 切分 key-value
			pairIndex := bytes.IndexByte(head, pairSplitter)
			key := string(pair[:pairIndex])
			// +1 也是为了跳掉分隔符
			value := string(pair[pairIndex+1:])
			metaMap[key] = value

			// 往前移动
			head = head[index+1:]
			index = bytes.IndexByte(head, splitter)
		}
		req.Meta = metaMap
	}
	req.Data = data[req.HeadLength:]
	return req
}

type Response struct {
	HeadLength uint32 `json:"head_length,omitempty"`
	BodyLength uint32 `json:"body_length,omitempty"`

	MessageId uint32 `json:"message_id,omitempty"`

	Version    byte `json:"version,omitempty"`
	Compressor byte `json:"compressor,omitempty"`
	Serializer byte `json:"serializer,omitempty"`

	Error []byte `json:"error,omitempty"`
	// 你要区分业务 error 还是非业务 error
	// BizError []byte // 代表的是业务返回的 error

	Data []byte `json:"data,omitempty"`
}

func (resp *Response) SetHeadLength() {
	resp.HeadLength = uint32(15 + len(resp.Error))
}

// 这里处理 Resp 我直接复制粘贴，是因为我觉得复制粘贴会使可读性更高

func EncodeResp(resp *Response) []byte {
	bs := make([]byte, resp.HeadLength+resp.BodyLength)

	cur := bs
	// 1. 写入 HeadLength，四个字节
	binary.BigEndian.PutUint32(cur[:4], resp.HeadLength)
	cur = cur[4:]
	// 2. 写入 BodyLength 四个字节
	binary.BigEndian.PutUint32(cur[:4], resp.BodyLength)
	cur = cur[4:]

	// 3. 写入 message id, 四个字节
	binary.BigEndian.PutUint32(cur[:4], resp.MessageId)
	cur = cur[4:]

	// 4. 写入 version，因为本身就是一个字节，所以不用进行编码了
	cur[0] = resp.Version
	cur = cur[1:]

	// 5. 写入压缩算法
	cur[0] = resp.Compressor
	cur = cur[1:]

	// 6. 写入序列化协议
	cur[0] = resp.Serializer
	cur = cur[1:]
	// 7. 写入 error
	copy(cur, resp.Error)
	cur = cur[len(resp.Error):]

	// 剩下的数据
	copy(cur, resp.Data)
	return bs
}

// DecodeResp 解析 Response
func DecodeResp(bs []byte) *Response {
	resp := &Response{}
	// 按照 EncodeReq 写下来
	// 1. 读取 HeadLength
	resp.HeadLength = binary.BigEndian.Uint32(bs[:4])
	// 2. 读取 BodyLength
	resp.BodyLength = binary.BigEndian.Uint32(bs[4:8])
	// 3. 读取 message id
	resp.MessageId = binary.BigEndian.Uint32(bs[8:12])
	// 4. 读取 Version
	resp.Version = bs[12]
	// 5. 读取压缩算法
	resp.Compressor = bs[13]
	// 6. 读取序列化协议
	resp.Serializer = bs[14]
	// 7. error 信息
	resp.Error = bs[15:resp.HeadLength]

	// 剩下的就是数据了
	resp.Data = bs[resp.HeadLength:]
	return resp
}
