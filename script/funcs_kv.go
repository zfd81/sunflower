package script

import (
	"strings"

	"github.com/zfd81/rock/conf"

	"github.com/zfd81/rock/meta"

	"github.com/zfd81/rock/meta/dai"

	"github.com/robertkrimen/otto"
)

func KvGet(process Processor) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		name_v := call.Argument(0)
		if name_v.IsUndefined() || name_v.IsNull() {
			return ErrorResult(call, "KVS name cannot be empty")
		}
		name := strings.TrimSpace(name_v.String()) //获取kvs名称
		if name == "" {
			return ErrorResult(call, "KVS name cannot be empty")
		}
		key_v := call.Argument(1)
		if key_v.IsUndefined() || key_v.IsNull() {
			return ErrorResult(call, "KV key cannot be empty")
		}
		key := strings.TrimSpace(key_v.String()) //获取kvs名称
		if key == "" {
			return ErrorResult(call, "KV key cannot be empty")
		}
		kv, err := dai.GetKV(process.GetNamespace(), meta.FormatPath(name)+"/"+key)
		if err != nil {
			return ErrorResult(call, err.Error())
		}
		var data interface{}
		if kv != nil {
			data = kv.Value
		}
		return Result(call, data)
	}
}

func KvSet(process Processor) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		name_v := call.Argument(0)
		if name_v.IsUndefined() || name_v.IsNull() {
			return ErrorResult(call, "KVS name cannot be empty")
		}
		name := strings.TrimSpace(name_v.String()) //获取kvs名称
		if name == "" {
			return ErrorResult(call, "KVS name cannot be empty")
		}
		key_v := call.Argument(1)
		if key_v.IsUndefined() || key_v.IsNull() {
			return ErrorResult(call, "KV key cannot be empty")
		}
		key := strings.TrimSpace(key_v.String()) //获取kvs名称
		if key == "" {
			return ErrorResult(call, "KV key cannot be empty")
		}
		value_v := call.Argument(2)
		if value_v.IsUndefined() || value_v.IsNull() {
			return ErrorResult(call, "KV value cannot be empty")
		}
		value, err := value_v.Export()
		if err != nil {
			return ErrorResult(call, err.Error())
		}
		ttl := conf.GetConfig().KVTTL
		ttl_v := call.Argument(3)
		if ttl_v.IsNumber() {
			if ttl, err = ttl_v.ToInteger(); err != nil {
				return ErrorResult(call, err.Error())
			}
		}
		kv := &meta.KV{
			Namespace: process.GetNamespace(),
			KvsName:   name,
			Key:       key,
			TTL:       ttl,
		}
		if err = kv.SetValue(value); err != nil {
			return ErrorResult(call, err.Error())
		}
		if err = dai.SetKV(kv); err != nil {
			return ErrorResult(call, err.Error())
		}
		return Result(call, nil)
	}
}