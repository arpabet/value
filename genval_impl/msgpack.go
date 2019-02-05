package genval_impl

const (

	mpPosFixIntMask 	byte = 0x80
	mpFixMapPrefix  	byte = 0x80
	mpFixArrayPrefix  	byte = 0x90
	mpFixStrPrefix    	byte = 0xa0

	mpNil          		byte = 0xc0
	mpNeverUsed    		byte = 0xc1
	mpFalse        		byte = 0xc2
	mpTrue         		byte = 0xc3

	mpBin8     			byte = 0xc4
	mpBin16    			byte = 0xc5
	mpBin32    			byte = 0xc6
	mpExt8     			byte = 0xc7
	mpExt16    			byte = 0xc8
	mpExt32    			byte = 0xc9

	mpFloat32   		byte = 0xca
	mpFloat64   		byte = 0xcb

	mpUint8        		byte = 0xcc
	mpUint16       		byte = 0xcd
	mpUint32       		byte = 0xce
	mpUint64       		byte = 0xcf

	mpInt8         		byte = 0xd0
	mpInt16        		byte = 0xd1
	mpInt32        		byte = 0xd2
	mpInt64        		byte = 0xd3

	mpFixExt1  			byte = 0xd4
	mpFixExt2  			byte = 0xd5
	mpFixExt4  			byte = 0xd6
	mpFixExt8  			byte = 0xd7
	mpFixExt16 			byte = 0xd8

	mpStr8  			byte = 0xd9
	mpStr16 			byte = 0xda
	mpStr32 			byte = 0xdb

	mpArray16 			byte = 0xdc
	mpArray32 			byte = 0xdd

	mpMap16 			byte = 0xde
	mpMap32 			byte = 0xdf

	mpNegFixIntPrefix 	byte = 0xe0

)

