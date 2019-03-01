package main

import (
 "fmt";	"encoding/base64"; "crypto/rand"; 
 "github.com/ztrue/tracerr"
 "math/big")

func genNonce(length uint8) {
 fmt.Println("FIXME: test that those numbers are sufficiently random!")
 for i:=0; i<50; i++ {
  str, err := genNonceInner(length)
  if err != nil {
   panic(err) }
  fmt.Println("Nonce1:",str) }}
 
// Original Python function I simplified it a little
/* def gen_nonce(length):
   """ Generates a random string of bytes, base64 encoded """
   if length < 1:
      return ''
   string=base64.b64encode(os.urandom(length),altchars=b'-_')
   b64len=4* (length // 3)
   if length%3 == 1:
      b64len+=2
   elif length%3 == 2:
      b64len+=3
   return string[0:b64len].decode() */

// Generates a random string of bytes, base64 encoded
// Inspired by the discussion in the https://github.com/joestump/python-oauth2/issues/9#
func genNonceInner(length uint8)	(res string, err error) {
 res = ""
	if length == 0 { 
  err = tracerr.New("Zero length of nonce is a nonsense!")
  return }
 maxx := big.NewInt(256);
 maxx = maxx.Exp(maxx,big.NewInt(int64(length)),nil)
 var nonce *big.Int;
 nonce, err = rand.Int(rand.Reader,maxx); 
 if err != nil {
  fmt.Println("Unable to generate a random link")
  return
 }
 nonceString := nonce.String()
 res = base64.RawURLEncoding.EncodeToString([]byte(nonceString))
 return
}

 