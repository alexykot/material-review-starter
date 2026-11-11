import {
  CreateBucketCommand,
  GetObjectCommand,
  HeadBucketCommand,
  PutObjectCommand,
  S3Client,
} from "@aws-sdk/client-s3";
import { getSignedUrl } from "@aws-sdk/s3-request-presigner";
import { env } from "./env.ts";

export const s3 = new S3Client({
  endpoint: env.s3.endpoint,
  region: env.s3.region,
  forcePathStyle: true,
  credentials: { accessKeyId: env.s3.accessKeyId, secretAccessKey: env.s3.secretAccessKey },
});

const bucket = env.s3.bucket;

export async function ensureBucket(): Promise<void> {
  try {
    await s3.send(new HeadBucketCommand({ Bucket: bucket }));
  } catch {
    await s3.send(new CreateBucketCommand({ Bucket: bucket }));
  }
}

export async function uploadObject(key: string, body: Buffer | Uint8Array | string, contentType: string): Promise<void> {
  await s3.send(new PutObjectCommand({ Bucket: bucket, Key: key, Body: body, ContentType: contentType }));
}

export async function getObject(key: string): Promise<{ body: Buffer; contentType: string | undefined }> {
  const result = await s3.send(new GetObjectCommand({ Bucket: bucket, Key: key }));
  const bytes = await result.Body!.transformToByteArray();
  return { body: Buffer.from(bytes), contentType: result.ContentType };
}

export async function getDownloadUrl(key: string, filename?: string, expiresInSeconds = 900): Promise<string> {
  const command = new GetObjectCommand({
    Bucket: bucket,
    Key: key,
    ResponseContentDisposition: filename ? `attachment; filename="${filename}"` : undefined,
  });
  return getSignedUrl(s3, command, { expiresIn: expiresInSeconds });
}
