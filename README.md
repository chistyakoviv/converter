# A Web Service for Converting Images and Videos to Modern Formats

The service utilizes [govips](https://github.com/davidbyttow/govip) and [FFmpeg](https://github.com/FFmpeg/FFmpeg) to convert media files to modern formats efficiently.

## Running the Server

The repository includes Dockerfiles to facilitate running the server. Clone the repository and execute the following command:

```bash
docker compose up
```

The application will be accessible at `http://localhost` on port 80.

## Usage

### Required Project Structure

To run the server, the following directories must exist:

- **migrations**: Contains migration files.
- **files**: Contains files to be converted.
- **config**: Contains configuration files.

### Configuration Files

Configuration files are divided into two types:

1. **Application Configuration**
2. **Default Conversion Configuration**

#### Application Configuration

| Option          | Range of Values   | Default Value | Required | Description                                                                 |
|-----------------|------------------|---------------|----------|-----------------------------------------------------------------------------|
| **env**         | local, dev, prod |               | Yes      | The environment the application is running in.                             |
| **Database**    |                  |               |          |                                                                             |
| dsn             |                  |               | Yes      | DSN string to connect to the PostgreSQL database.                          |
| **HTTP Server** |                  |               |          |                                                                             |
| address         |                  | 0.0.0.0:80    | No       | Address where the web server will start.                                   |
| read timeout    |                  |               | Yes      | Maximum duration for reading the entire request, including the body.       |
| write timeout   |                  |               | Yes      | Maximum duration for writing the response to the client.                   |
| idle timeout    |                  |               | Yes      | Maximum duration for keeping an idle connection open.                      |
| **Task**        |                  |               |          |                                                                             |
| check timeout   |                  | 5m            | No       | Interval to check for new tasks available for execution.                   |
| **Image**       |                  |               |          |                                                                             |
| threads         |                  | 1             | No       | Number of threads for converting images.                                   |
| **Video**       |                  |               |          |                                                                             |
| threads         |                  | 1             | No       | Number of threads for converting videos.                                   |

The application configuration can be provided via the `CONFIG_PATH` environment variable. If `CONFIG_PATH` is not set, all options will be read from individual environment variables:

| Option         | Environment Variable |
|----------------|-----------------------|
| app env        | ENV                   |
| database dsn   | POSTGRES_DSN          |
| http address   | ADDRESS               |
| http read timeout | READ_TIMEOUT       |
| http write timeout | WRITE_TIMEOUT     |
| http idle timeout  | IDLE_TIMEOUT      |
| task check timeout | TASK_CHECK_TIMEOUT |
| image threads  | IMAGE_THREADS         |
| video threads  | VIDEO_THREADS         |

#### Default Conversion Configuration

The default configuration is merged with the configuration received from a request, allowing overrides via request fields.

| Option          | Description                                                                         |
|-----------------|-------------------------------------------------------------------------------------|
| **Image Options** |                                                                                   |
| formats         | Array of key-value pairs passed to govips.                                          |
| **Video Options** |                                                                                   |
| formats         | Array of key-value pairs passed to FFmpeg.                                         |

Examples for both configurations can be found in the `config` directory.

### API Endpoints

The service provides three endpoints:

- `POST /convert`: Enqueue a file for conversion.
- `POST /delete`: Delete converted files for a specified file.
- `POST /scan`: Scan the `files` directory and enqueue found files for conversion.

#### Conversion Request

| Field Name     | Description                                                                   |
|----------------|-------------------------------------------------------------------------------|
| path           | Path to a file that should exist in the `files` directory for successful conversion. |
| convert_to     | Array of conversion options.                                                 |

**`convert_to` Description**

| Field Name     | Description                                                                   |
|----------------|-------------------------------------------------------------------------------|
| ext            | Extension for the output file.                                               |
| conv_conf      | Conversion configuration options (refer to [govips](https://github.com/davidbyttow/govip) and [FFmpeg](https://github.com/FFmpeg/FFmpeg) for details). |
| optional       | Optional settings.                                                           |

**`optional` Description**

| Field Name         | Description                                                                 |
|--------------------|-----------------------------------------------------------------------------|
| replace_orig_ext   | If true, replaces the original extension with the `ext` field value.       |
| suffix             | Adds a suffix to differentiate files with the same output extension.       |

**Supported Conversions**

| Extension           | Supported Conversion Formats                                              |
|---------------------|---------------------------------------------------------------------------|
| jpg, jpeg, png      | jpg, jpeg, png, webp, avif                                                |
| mp4, webm           | mp4, webm                                                                |

**Example: Video Conversion Request**

```json
{
  "path": "/files/videos/video.mp4",
  "convert_to": [
    {
      "ext": "webm",
      "conv_conf": {
        "crf": 35
      },
      "optional": {
        "replace_orig_ext": true,
        "suffix": ".vp9"
      }
    },
    {
      "ext": "webm",
      "conv_conf": {
        "crf": 50
      },
      "optional": {
        "replace_orig_ext": true,
        "suffix": ".av1"
      }
    }
  ]
}
```

**Example: Image Conversion Request**

If `convert_to` is empty, the default conversion configuration will be used.

```json
{
  "path": "/files/images/image.jpg"
}
```

#### Deletion Request

| Field Name     | Description                                                                   |
|----------------|-------------------------------------------------------------------------------|
| path           | Path to a file for which all converted files should be deleted. The original file need not exist. |

#### Scan Request

The scan request does not require parameters.

## Using the Package in Your Project

1. Create a `main` package with the following code:

    ```go
    package main

    import (
        "context"

        "github.com/chistyakoviv/converter/app"
    )

    func main() {
        ctx := context.Background()
        a := app.NewApp(ctx)
        a.Run(ctx)
    }
    ```

2. Copy the `migrations` directory from the repository.

3. Create a `config` directory and add default settings for your project.

4. Create a `files` directory and add files you need to convert.

