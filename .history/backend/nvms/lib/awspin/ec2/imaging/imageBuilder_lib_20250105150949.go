func (c *Client) buildEndpoint() (string, error) {
    u, err := url.Parse(c.endpointURL)
    if err != nil {
        return "", fmt.Errorf("failed to parse endpoint: %w", err)
    }
    return u.String(), nil
}

func (c *Client) newRequest(ctx context.Context, method string, path string, body []byte) (*http.Request, error) {
    endpoint, err := c.buildEndpoint()
    if err != nil {
        return nil, err
    }
    
    u, err := url.Parse(endpoint + path)
    if err != nil {
        return nil, err
    }

    // Create request with body
    req, err := http.NewRequestWithContext(ctx, method, u.String(), bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    // Set required headers
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("User-Agent", "byteport")

    // Add AWS v4 signature
    var awsDate aws.AwsDate
    awsDate.Time = time.Now()
    
    // Create canonical request
    payloadHash := aws.GetSHA256Hash(body)
    
    canonicalURI := path
    canonicalQueryString := ""
    canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\n",
        req.Header.Get("Content-Type"),
        u.Host)
    signedHeaders := "content-type;host"
    
    canonicalRequest := strings.Join([]string{
        method,
        canonicalURI,
        canonicalQueryString,
        canonicalHeaders,
        signedHeaders,
        payloadHash,
    }, "\n")

    // Create string to sign
    credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request",
        awsDate.GetDate(),
        c.config.Region,
        c.config.Service)

    stringToSign := strings.Join([]string{
        "AWS4-HMAC-SHA256",
        awsDate.GetTime(),
        credentialScope,
        aws.GetSHA256Hash([]byte(canonicalRequest)),
    }, "\n")

    // Calculate signature
    dateKey := aws.HmacSHA256([]byte("AWS4"+c.config.SecretAccessKey), []byte(awsDate.GetDate()))
    regionKey := aws.HmacSHA256(dateKey, []byte(c.config.Region))
    serviceKey := aws.HmacSHA256(regionKey, []byte(c.config.Service))
    signingKey := aws.HmacSHA256(serviceKey, []byte("aws4_request"))
    signature := hex.EncodeToString(aws.HmacSHA256(signingKey, []byte(stringToSign)))

    // Add authorization header
    authHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
        c.config.AccessKeyId,
        credentialScope,
        signedHeaders,
        signature)
    
    req.Header.Set("Authorization", authHeader)
    req.Header.Set("X-Amz-Date", awsDate.GetTime())
    
    if c.config.SessionToken != "" {
        req.Header.Set("X-Amz-Security-Token", c.config.SessionToken)
    }

    return req, nil
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
    resp, err := spinhttp.Send(req)
    if err != nil {
        return nil, fmt.Errorf("failed to send request: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        var errorResponse struct {
            Message string `json:"message"`
            Type    string `json:"__type"`
        }
        if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
            return nil, fmt.Errorf("failed to parse error response: %w", err)
        }
        return nil, fmt.Errorf("%s: %s", errorResponse.Type, errorResponse.Message)
    }

    return resp, nil
}

func (c *Client) CreateInfrastructureConfiguration(ctx context.Context, name string, instProfName string) (*CreateInfrastructureConfigurationResponse, error) {
    reqBody := CreateInfrastructureConfigurationRequest{
        ClientToken: time.Now().String(),
        Description: "Infrastructure Configuration for Image Builder",
        InstanceProfileName: instProfName,
        Name: name,
    }

    body, err := json.Marshal(reqBody)
    if err != nil {
        return nil, err
    }

    req, err := c.newRequest(ctx, "PUT", "/CreateInfrastructureConfiguration", body)
    if err != nil {
        return nil, err
    }

    resp, err := c.do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result CreateInfrastructureConfigurationResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    return &result, nil
}