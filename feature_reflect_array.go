package jsoniter

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"
)

func decoderOfArray(cfg *frozenConfig, typ reflect.Type) (ValDecoder, error) {
	decoder, err := decoderOfType(cfg, typ.Elem())
	if err != nil {
		return nil, err
	}
	return &arrayDecoder{typ, typ.Elem(), decoder}, nil
}

func encoderOfArray(cfg *frozenConfig, typ reflect.Type) (ValEncoder, error) {
	encoder, err := encoderOfType(cfg, typ.Elem())
	if err != nil {
		return nil, err
	}
	if typ.Elem().Kind() == reflect.Map {
		encoder = &optionalEncoder{encoder}
	}
	return &arrayEncoder{typ, typ.Elem(), encoder}, nil
}

type arrayEncoder struct {
	arrayType   reflect.Type
	elemType    reflect.Type
	elemEncoder ValEncoder
}

func (encoder *arrayEncoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	if ptr == nil {
		stream.WriteNil()
		return
	}
	stream.WriteArrayStart()
	elemPtr := uintptr(ptr)
	encoder.elemEncoder.Encode(unsafe.Pointer(elemPtr), stream)
	for i := 1; i < encoder.arrayType.Len(); i++ {
		stream.WriteMore()
		elemPtr += encoder.elemType.Size()
		encoder.elemEncoder.Encode(unsafe.Pointer(elemPtr), stream)
	}
	stream.WriteArrayEnd()
	if stream.Error != nil && stream.Error != io.EOF {
		stream.Error = fmt.Errorf("%v: %s", encoder.arrayType, stream.Error.Error())
	}
}

func (encoder *arrayEncoder) EncodeInterface(val interface{}, stream *Stream) {
	WriteToStream(val, stream, encoder)
}

func (encoder *arrayEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return false
}

type arrayDecoder struct {
	arrayType   reflect.Type
	elemType    reflect.Type
	elemDecoder ValDecoder
}

func (decoder *arrayDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	decoder.doDecode(ptr, iter)
	if iter.Error != nil && iter.Error != io.EOF {
		iter.Error = fmt.Errorf("%v: %s", decoder.arrayType, iter.Error.Error())
	}
}

func (decoder *arrayDecoder) doDecode(ptr unsafe.Pointer, iter *Iterator) {
	offset := uintptr(0)
	for ; iter.ReadArray(); offset += decoder.elemType.Size() {
		if offset < decoder.arrayType.Size() {
			decoder.elemDecoder.Decode(unsafe.Pointer(uintptr(ptr)+offset), iter)
		} else {
			iter.Skip()
		}
	}
}
