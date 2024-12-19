import * as util from "./awsutil";

// Function to create an S3 bucket

import { AwsCreds, Project, signRequest } from "./awsutil";

export async function createBucket(
	bucketName: string,
	region: string,
	creds: AwsCreds
): Promise<void> {
	const method = "PUT";
	const service = "s3";
	const url = `https://${bucketName}.s3.${region}.amazonaws.com/`;

	const headers = await signRequest(
		method,
		url,
		region,
		service,
		"",
		creds.accessKeyId,
		creds.secretAccessKey,
		{
			"x-amz-content-sha256": await util.sha256(""),
		}
	);

	const response = await fetch(url, {
		method,
		headers, // Use the plain object
	});

	if (response.status !== 200 && response.status !== 409) {
		// 200 OK or 409 Conflict (bucket already exists)
		const errorText = await response.text();
		throw new Error(`Failed to create bucket: ${response.status} ${errorText}`);
	}
}

export async function putObject(
	bucketName: string,
	objectKey: string,
	region: string,
	creds: AwsCreds,
	body: Uint8Array
): Promise<void> {
	const method = "PUT";
	const service = "s3";
	const url = `https://${bucketName}.s3.${region}.amazonaws.com/${encodeURIComponent(
		objectKey
	)}`;

	const headers = await signRequest(
		method,
		url,
		region,
		service,
		body,
		creds.accessKeyId,
		creds.secretAccessKey,
		{
			"Content-Type": "application/octet-stream",
			"x-amz-content-sha256": await util.sha256(body),
		}
	);

	const response = await fetch(url, {
		method,
		headers, // Use the plain object
		body,
	});

	if (response.status !== 200) {
		const errorText = await response.text();
		throw new Error(`Failed to upload object: ${response.status} ${errorText}`);
	}
}
