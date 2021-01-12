package devportalservice

const testDevicesResponseBody = `{
   "test_devices":[
      {
         "id":24,
         "user_id":4,
         "device_identifier":"asdf12345ad9b298cb9a9f28555c49573d8bc322",
         "title":"iPhone 6",
         "created_at":"2015-03-13T16:16:13.665Z",
         "updated_at":"2015-03-13T16:16:13.665Z",
         "device_type":"ios"
      },
      {
         "id":28,
         "user_id":4,
         "device_identifier":"asdf12341e73b76df6e99d0d713133c3e078418f",
         "title":"iPad mini 2 (Wi-Fi)",
         "created_at":"2015-03-19T13:25:43.487Z",
         "updated_at":"2015-03-19T13:25:43.487Z",
         "device_type":"ios"
	  }
	]
}
`

const testAppleDevConnSession = `---
- !ruby/object:HTTP::Cookie
  name: DES58b0eba556d80ed2b98707e15ffafd344
  value: HSARMTKNSRVTWFlaFrGQTmfmFBwJuiX/aaaaaaaaa+A7FbJa4V8MmWijnJknnX06ME0KrI9V8vFg==SRVT
  domain: idmsa.apple.com
  for_domain: true
  path: "/"

- !ruby/object:HTTP::Cookie
  name: myacinfo
  value: DAWTKNV26a0a6db3ae43acd203d0d03e8bc45000cd4bdc668e90953f22ca3b36eaab0e18634660a10cf28cc65d8ddf633c017de09477dfb18c8a3d6961f96cbbf064be616e80cee62d3d7f39a485bf826377c5b5dbbfc4a97dcdb462052db73a3a1d9b4a325d5bdd496190b3088878cecce17e4d6db9230e0575cfbe7a8754d1de0c937080ef84569b6e4a75237c2ec01cf07db060a11d92e7220707dd00a2a565ee9e06074d8efa6a1b7f83db3e1b2acdafb5fc0708443e77e6d71e168ae2a83b848122264b2da5cadfd9e451f9fe3f6eebc71904d4bc36acc528cc2a844d4f2eb527649a69523756ec9955457f704c28a3b6b9f97d6df900bd60044d5bc50408260f096954f03c53c16ac40a796dc439b859f882a50390b1c7517a9f4479fb1ce9ba2db241d6b8f2eb127c46ef96e0ccccccccc
  domain: apple.com
  for_domain: true
  path: "/"

`

const testAppleDevConnDataJSON = `{
    "apple_id": "example@example.io",
    "password": "highSecurityPassword",
    "connection_expiry_date": "2019-04-06T12:04:59.000Z",
    "session_cookies": {
        "https://idmsa.apple.com": [
            {
                "name": "DES58b0eba556d80ed2b98707e15ffafd344",
                "path": "/",
                "value": "HSARMTKNSRVTWFlaFrGQTmfmFBwJuiX/aaaaaaaaa+A7FbJa4V8MmWijnJknnX06ME0KrI9V8vFg==SRVT",
                "domain": "idmsa.apple.com",
                "secure": true,
                "expires": "2019-04-06T12:04:59Z",
                "max_age": 2592000,
                "httponly": true
            },
            {
                "name": "myacinfo",
                "path": "/",
                "value": "DAWTKNV26a0a6db3ae43acd203d0d03e8bc45000cd4bdc668e90953f22ca3b36eaab0e18634660a10cf28cc65d8ddf633c017de09477dfb18c8a3d6961f96cbbf064be616e80cee62d3d7f39a485bf826377c5b5dbbfc4a97dcdb462052db73a3a1d9b4a325d5bdd496190b3088878cecce17e4d6db9230e0575cfbe7a8754d1de0c937080ef84569b6e4a75237c2ec01cf07db060a11d92e7220707dd00a2a565ee9e06074d8efa6a1b7f83db3e1b2acdafb5fc0708443e77e6d71e168ae2a83b848122264b2da5cadfd9e451f9fe3f6eebc71904d4bc36acc528cc2a844d4f2eb527649a69523756ec9955457f704c28a3b6b9f97d6df900bd60044d5bc50408260f096954f03c53c16ac40a796dc439b859f882a50390b1c7517a9f4479fb1ce9ba2db241d6b8f2eb127c46ef96e0ccccccccc",
                "domain": "apple.com",
                "secure": true,
                "httponly": true
            }
        ]
    },
    "test_devices": [
        {
            "id": 8414,
            "user_id": 52411,
            "device_identifier": "1b78ac4bad2e8911139287ac5dd152fbe86eb2b9",
            "title": "iPhone 7",
            "created_at": "2018-08-30T09:09:36.332Z",
            "updated_at": "2018-08-30T09:09:36.332Z",
            "device_type": "ios"
        }
    ],
    "default_team_id": null
}`

var testAppleDevConnData = AppleDeveloperConnection{
	AppleID:              "example@example.io",
	Password:             "highSecurityPassword",
	ConnectionExpiryDate: "2019-04-06T12:04:59.000Z",
	SessionCookies: map[string][]cookie{
		"https://idmsa.apple.com": {
			{
				Name:     "DES58b0eba556d80ed2b98707e15ffafd344",
				Path:     "/",
				Value:    "HSARMTKNSRVTWFlaFrGQTmfmFBwJuiX/aaaaaaaaa+A7FbJa4V8MmWijnJknnX06ME0KrI9V8vFg==SRVT",
				Domain:   "idmsa.apple.com",
				Secure:   true,
				Expires:  "2019-04-06T12:04:59Z",
				MaxAge:   2592000,
				Httponly: true,
			},
			{
				Name:     "myacinfo",
				Path:     "/",
				Value:    "DAWTKNV26a0a6db3ae43acd203d0d03e8bc45000cd4bdc668e90953f22ca3b36eaab0e18634660a10cf28cc65d8ddf633c017de09477dfb18c8a3d6961f96cbbf064be616e80cee62d3d7f39a485bf826377c5b5dbbfc4a97dcdb462052db73a3a1d9b4a325d5bdd496190b3088878cecce17e4d6db9230e0575cfbe7a8754d1de0c937080ef84569b6e4a75237c2ec01cf07db060a11d92e7220707dd00a2a565ee9e06074d8efa6a1b7f83db3e1b2acdafb5fc0708443e77e6d71e168ae2a83b848122264b2da5cadfd9e451f9fe3f6eebc71904d4bc36acc528cc2a844d4f2eb527649a69523756ec9955457f704c28a3b6b9f97d6df900bd60044d5bc50408260f096954f03c53c16ac40a796dc439b859f882a50390b1c7517a9f4479fb1ce9ba2db241d6b8f2eb127c46ef96e0ccccccccc",
				Domain:   "apple.com",
				Secure:   true,
				Httponly: true,
			},
		},
	},
}
