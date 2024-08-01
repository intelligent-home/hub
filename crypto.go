package hub

import "github.com/golang-module/dongle"

// Encrypt mqtt 消息编码
// 使用 3DES 加密输入字符串,然后用base64编码,返回字符串
func Encrypt(plainText []byte) ([]byte, error) {
	cipher := dongle.NewCipher()
	cipher.SetMode(dongle.CBC)      // CBC, CFB, CTR, ECB, OFB
	cipher.SetPadding(dongle.Zero)  // No、Empty、Zero、PKCS5、PKCS7、AnsiX923、ISO97971
	cipher.SetKey(MQTTSecretKey)    // key 长度必须是 24
	cipher.SetIV(MQTTSecretKey[:8]) // iv 长度必须是 8
	// 对字符串进行 3des 加密，输出经过 base64 编码的字符串
	enc := dongle.Encrypt.FromBytes(plainText).By3Des(cipher)
	if enc.Error != nil {
		return nil, enc.Error
	}
	return enc.ToBase64Bytes(), nil
}

// Decrypt mqtt 消息解码
func Decrypt(cipherText []byte) ([]byte, error) {
	cipher := dongle.NewCipher()
	cipher.SetMode(dongle.CBC)      // CBC, CFB, CTR, ECB, OFB
	cipher.SetPadding(dongle.Zero)  // No、Empty、Zero、PKCS5、PKCS7、AnsiX923、ISO97971
	cipher.SetKey(MQTTSecretKey)    // key 长度必须是 24
	cipher.SetIV(MQTTSecretKey[:8]) // iv 长度必须是 8
	// 对经过 base64 编码的字符串进行 3des 解密，输出字符串
	dec := dongle.Decrypt.FromBase64Bytes(cipherText).By3Des(cipher)
	if dec.Error != nil {
		return nil, dec.Error
	}
	return dec.ToBytes(), nil
}
