package ec2

func(c *Client) CreateInstanceProfile(ctx context.Context, name string) (*CreateInstanceProfileResponse, error) {
	params := map[string]string{
		"Action": "CreateInstanceProfile",
		"InstanceProfileName": name,
	}
	req, err := c.newRequest(ctx, "POST", params, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result CreateInstanceProfileResponse
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}