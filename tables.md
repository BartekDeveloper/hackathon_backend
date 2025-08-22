User

Table Name: user
Field Name	Type	Key	Description
id	string	
Unique identifier for each user
name	string	-	User's chosen display name
email	string	-	User's email address for communication and login
emailVerified	boolean	-	Whether the user's email is verified
image	string	
User's image url
createdAt	Date	-	Timestamp of when the user account was created
updatedAt	Date	-	Timestamp of the last update to the user's information
Session

Table Name: session
Field Name	Type	Key	Description
id	string	
Unique identifier for each session
userId	string	
The ID of the user
token	string	-	The unique session token
expiresAt	Date	-	The time when the session expires
ipAddress	string	
The IP address of the device
userAgent	string	
The user agent information of the device
createdAt	Date	-	Timestamp of when the session was created
updatedAt	Date	-	Timestamp of when the session was updated
Account

Table Name: account
Field Name	Type	Key	Description
id	string	
Unique identifier for each account
userId	string	
The ID of the user
accountId	string	-	The ID of the account as provided by the SSO or equal to userId for credential accounts
providerId	string	-	The ID of the provider
accessToken	string	
The access token of the account. Returned by the provider
refreshToken	string	
The refresh token of the account. Returned by the provider
accessTokenExpiresAt	Date	
The time when the access token expires
refreshTokenExpiresAt	Date	
The time when the refresh token expires
scope	string	
The scope of the account. Returned by the provider
idToken	string	
The ID token returned from the provider
password	string	
The password of the account. Mainly used for email and password authentication
createdAt	Date	-	Timestamp of when the account was created
updatedAt	Date	-	Timestamp of when the account was updated
Verification

Table Name: verification
Field Name	Type	Key	Description
id	string	
Unique identifier for each verification
identifier	string	-	The identifier for the verification request
value	string	-	The value to be verified
expiresAt	Date	-	The time when the verification request expires
createdAt	Date	-	Timestamp of when the verification request was created
updatedAt	Date	-	Timestamp of when the verification request was updated