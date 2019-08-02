package awsbase

import (
	"testing"
)

/*
Might be worth looking into a few mock configs
StaticProvider
SharedFileProvider

TODO: Validate that session contains the specified configuration overrides.
*/

func TestGetSessionOptions(t *testing.T) {
	tt := []struct {
		desc     string
		config   *Config
		hasError bool
	}{
		{"UnconfiguredConfig",
			&Config{},
			true,
		},
		{"ConfigWithCredentials",
			&Config{AccessKey: "MockAccessKey", SecretKey: "MockSecretKey"},
			false,
		},
		{"ConfigWithAllSupportedOptions",
			&Config{AccessKey: "MockAccessKey", SecretKey: "MockSecretKey", Insecure: true, DebugLogging: true},
			false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			opts, err := GetSessionOptions(tc.config)
			if err != nil && tc.hasError == false {
				t.Fatalf("GetSessionOptions(c) resulted in an error %s", err)
			}

			if opts == nil && tc.hasError == false {
				t.Error("GetSessionOptions(...) resulted in a nil set of options")
			}
		})

	}
}

func TestGetSession(t *testing.T) {
	_, err := GetSession(&Config{})
	if err == nil {
		t.Fatal("GetSession(&Config{}) with an empty config should result in an error but got nil")
	}

	/* Need a test case for getting a session that assumes a role

	Setting the AssumeRoleARN field triggers a call to the AWS API in the GetCredentials
	function that uses a raw AWS config client that does not contain any of the endpoint overrides
	being provided to the awsbase.Config object. I left a TODO in the code to determine if we need to honor
	any custom endpoints. Or if we should any calls for assuming a role should always use the default API endpoints.

	AssumeRoleARN:        "arn:aws:iam::222222222222:user/Alice",
	*/
	sess, err := GetSession(&Config{
		AccessKey:            "MockAccessKey",
		SecretKey:            "MockSecretKey",
		SkipCredsValidation:  true,
		SkipMetadataApiCheck: true,
		MaxRetries:           6,
		UserAgentProducts:    []*UserAgentProduct{{}},
	})
	if err != nil {
		t.Fatalf("GetSession(&Config{...}) should return a valid session, but got the error %s", err)
	}

	if sess == nil {
		t.Error("GetSession(...) resulted in a nil session")
	}
}

func TestGetSessionWithAccountIDAndPartition(t *testing.T) {
	ts := MockAwsApiServer("STS", []*MockEndpoint{
		{
			Request:  &MockRequest{"POST", "/", "Action=GetCallerIdentity&Version=2011-06-15"},
			Response: &MockResponse{200, stsResponse_GetCallerIdentity_valid, "text/xml"},
		},
	})
	defer ts.Close()

	tt := []struct {
		desc              string
		config            *Config
		expectedAcctID    string
		expectedPartition string
	}{
		//{"AssumeRoleARN_Config", &Config{AccessKey: "MockAccessKey", SecretKey: "MockSecretKey", AssumeRoleARN: "arn:aws:iam::222222222222:user/Alice", SkipMetadataApiCheck: true}, "222222222222", "aws"},
		{"StandardProvider_Config", &Config{AccessKey: "MockAccessKey", SecretKey: "MockSecretKey", Region: "us-west-2", UserAgentProducts: []*UserAgentProduct{{}}, StsEndpoint: ts.URL}, "222222222222", "aws"},
		{"SkipCredsValidation_Config", &Config{AccessKey: "MockAccessKey", SecretKey: "MockSecretKey", Region: "us-west-2", SkipCredsValidation: true, UserAgentProducts: []*UserAgentProduct{{}}, StsEndpoint: ts.URL}, "222222222222", "aws"},
		{"SkipRequestingAccountId_Config", &Config{AccessKey: "MockAccessKey", SecretKey: "MockSecretKey", Region: "us-west-2", SkipCredsValidation: true, SkipRequestingAccountId: true, UserAgentProducts: []*UserAgentProduct{{}}, StsEndpoint: ts.URL}, "", "aws"},
	}

	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			sess, acctID, part, err := GetSessionWithAccountIDAndPartition(tc.config)
			if err != nil {
				t.Fatalf("GetSessionWithAccountIDAndPartition(&Config{...}) should return a valid session, but got the error %s", err)
			}

			if sess == nil {
				t.Error("GetSession(c) resulted in a nil session")
			}

			if acctID != tc.expectedAcctID {
				t.Errorf("GetSession(c) returned an incorrect AWS account ID, expected %q but got %q", tc.expectedAcctID, acctID)
			}

			if part != tc.expectedPartition {
				t.Errorf("GetSession(c) returned an incorrect AWS partition, expected %q but got %q", tc.expectedPartition, part)
			}
		})
	}
}
