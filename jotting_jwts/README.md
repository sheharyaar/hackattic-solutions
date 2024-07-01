# JWTs

- broken in 3 parts (all **RawURLEncoded** (without padding bytes)), seperated by '.':
  - header
  - payload
  - signature
- header and payload are JSON objects
- header:
  - alg: algorithm used to sign the token
  - typ: type of token
- payload:
  - iss: issuer of the token
  - sub: subject of the token
  - aud: audience of the token
  - exp: expiration time
  - nbf: not before
  - iat: issued at
  - jti: JWT ID
  - custom claims
- signature (any algorithm as specified in header):
  ```
  RawURLEncoded(
    hmacSHA256(
        base64UrlEncode(header) + '.' + base64UrlEncode(payload),
        secret
    )
  )
  ```