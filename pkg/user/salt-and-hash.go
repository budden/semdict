package user

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"

	"golang.org/x/crypto/pbkdf2"

	// "github.com/ztrue/tracerr";
	"math/big"
)

// Оригинальная функция Python, я немного упростил её
/* def gen_nonce(length):
""" Генерирует случайную строку байтов в кодировке base64 """
if length < 1:
  return ''
string=base64.b64encode(os.urandom(length),altchars=b'-_')
b64len=4* (length // 3)
if length%3 == 1:
 b64len+=2
elif length%3 == 2:
 b64len+=3
return string[0:b64len].decode() */

func randomBytes(length uint8) []byte {
	if length == 0 {
		log.Fatal("Случайный массив длины 0 не является случайным!")
	}
	maxx := big.NewInt(256)
	maxx = maxx.Exp(maxx, big.NewInt(int64(length)), nil)
	var randomInt *big.Int
	randomInt, err := rand.Int(rand.Reader, maxx)
	if err != nil {
		msg := fmt.Sprintf(
			"Невозможно сгенерировать случайные байты, ошибка = %v", err)
		log.Fatal(msg)
	}
	return randomInt.Bytes()
}

// GenNonce Генерирует случайную строку байтов в кодировке base64
// Под впечатлением от дискуссии в https://github.com/joestump/python-oauth2/issues/9#
func GenNonce(length uint8) string {
	nonceBytes := randomBytes(length)
	nonceString := string(nonceBytes)
	res := base64.RawURLEncoding.EncodeToString([]byte(nonceString))
	return res
}

const saltBytes = 16

// SaltAndHashPassword генерирует динамическую соль, хэш и возвращает оба параметра
// https://habr.com/ru/post/145648/
func SaltAndHashPassword(password string) (saltBase64, dkBase64 string) {
	salt := randomBytes(saltBytes)
	dk := pbkdf2.Key([]byte(password), salt, 4096, 32, sha1.New)
	saltBase64 = base64.RawURLEncoding.EncodeToString(salt)
	dkBase64 = base64.RawURLEncoding.EncodeToString(dk)
	return
}

// CheckPasswordAgainstSaltAndHash сопоставляет пароль с парой хэш/соль
func CheckPasswordAgainstSaltAndHash(password, saltBase64, dkBase64 string) bool {
	salt, err := base64.RawURLEncoding.DecodeString(saltBase64)
	if err != nil {
		return false
	}
	dk := pbkdf2.Key([]byte(password), salt, 4096, 32, sha1.New)
	dkBase642 := base64.RawURLEncoding.EncodeToString(dk)
	return (dkBase64 == dkBase642)
}

/*


// http://security.stackexchange.com/questions/110084/parameters-for-pbkdf2-for-password-hashing
type HashingConfig struct {
hashBytes int // размер генерируемого хэша (выбирается в соответствии с выбранным алгоритмом)
 saltBytes int // размер соли : большая соль означает, что хэшированные пароли более устойчивы к радужной таблице
 iterations int // настроить так, чтобы хэширование пароля занимало около 1 секунды
 algo string
 encoding string // hex читается лучше, но base64 короче
}

var config = HashingConfig{
 hashBytes  : 64,
 saltBytes  : 16,
 iterations : 220000,
 algo       :"sha512",
 encoding   : "base64" };



 /**
  * Проверка пароля с помощью асинхронной функции pbkdf2 (выведение ключа) Node.
  *
  * Принимает хэш и соль, сгенерированные hashPassword, и возвращает,
  * соответствует ли хэш паролю (как разрешённое обещание).
*/
/* function verifyPassword(password, hashframe) {
    // decode and extract hashing parameters
    hashframe = Buffer.from(hashframe, config.encoding);
    var saltBytes  = hashframe.readUInt32BE(0);
    var hashBytes  = hashframe.length - saltBytes - 8;
    var iterations = hashframe.readUInt32BE(4);
    var salt = hashframe.slice(8, saltBytes + 8);
    var hash = hashframe.slice(8 + saltBytes, saltBytes + hashBytes + 8);
    // verify the salt and hash against the password
    return crypto.pbkdf2Async(password, salt, iterations, hashBytes, config.algo)
        .then(function(verify) {
            if (verify.equals(hash)) return Promise.resolve(true);
            return Promise.resolve(false) ;
        })
}

exports.hashPassword = hashPassword;
exports.verifyPassword = verifyPassword;
*/
// используется для тестирования
/*
 console.time("hash");
 hashPassword("abc")
     .then(function(hash) {
         console.log("hashframe", hash.length, hash);
         console.timeEnd("hash");
         return verifyPassword("abc", hash);
     })
     .then(function()     { console.log("password correct");})
     .catch(function(err) { console.log("err", err);})
*/
