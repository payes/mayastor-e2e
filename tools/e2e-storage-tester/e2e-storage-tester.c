/***
 This program is a simple I/O tester with functionality to retry a failed
 I/O operation until a timeout expires.
 It writes and/or reads 4096 bytes at a time, starting at offset 0, using
 direct I/O and can write and verify a predictable byte pattern in each
 4096 byte block. If both writing and reading/verifying, it writes the
 patterns first and then reads/verifies the same data.
 The total number of blocks accessed by the last operation is written to
 stdout at the end of the test.
 See  help() for a description of the parameters.
***/

#define _GNU_SOURCE

#include <assert.h>
#include <errno.h>
#include <fcntl.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/types.h>
#include <unistd.h>

#define BLOCKSIZE (4096)
#define ALIGNMENT BLOCKSIZE

typedef enum {
    EREAD,
    EWRITE
} EOPERATION;

void help(const char *error_message, const char * param) {
	const char * usage_message =
    "usage: e2e-storage-tester <options> <device>\n"
    "\t -r read\n"
    "\t -t <value> seconds allowed for retried I/O attempts (default 50, 0 means do not retry)\n"
    "\t -v validate\n"
    "\t -w write\n"
    "\t -n <value> number of blocks (default 1)\n";
    if (param[0] != '\0') {
        fprintf(stderr, "error with parameter %s: ", param);
    }
    if (error_message[0] != '\0') {
        fprintf(stderr, "%s, ", error_message);
    }
    fprintf(stderr, "%s\n", usage_message);
    exit(-1);
}

/* write a predictable pattern to the block that depends on the block location */
const size_t BLOCKNOLENGTH = 8;
void blockPattern(int blockNumber, char * buffer, int bytes)
{
    assert(bytes >= BLOCKNOLENGTH && buffer != NULL && bytes <= BLOCKSIZE);

    for (int i = 0; i < BLOCKNOLENGTH; i++) {
        buffer[i] = (char)(blockNumber % 256);
        blockNumber /= 256;
    }
    memset(&(buffer[BLOCKNOLENGTH]), 0, bytes - BLOCKNOLENGTH);
}

char pattern_buffer[BLOCKSIZE] __attribute__((__aligned__(ALIGNMENT)));
char read_buffer[BLOCKSIZE] __attribute__((__aligned__(ALIGNMENT)));

/* attempt to read or write a block, with retries allowed up to a given timeout */
int ioAttempt(
    int fd,
    char * buffer,
    EOPERATION op,
    int block_no,
    int retryIntervalSecs,
    int retryTimeoutSecs)
{
    assert(buffer != NULL);

    int err = 0;
    int elapsedSecs = 0;
    const char * action_text = "";

    if (op == EWRITE) {
        blockPattern(block_no, buffer, BLOCKSIZE);
        action_text = "write";
    } else {
        action_text = "read";
    }
    for (;;)
    {
        err = lseek(fd, (int64_t)(block_no * BLOCKSIZE), 0);
        if (err < 0) {
            fprintf(stderr, " could not seek at block %d, error: %s\n", block_no, strerror(errno));
        } else {
            int64_t num_bytes;
            if (op == EWRITE) {
                num_bytes = write(fd, buffer, BLOCKSIZE);
            } else {
                num_bytes = read(fd, buffer, BLOCKSIZE);
            }
            if (num_bytes != BLOCKSIZE) {
                fprintf(stderr, "could not %s at block %d, wanted %d, got %ld, error: %s\n", action_text, block_no, BLOCKSIZE, num_bytes, strerror(errno));
                fprintf(stderr, "attempted seconds %d vs timeout %d\n", elapsedSecs, retryTimeoutSecs);
                err = -1;
            } else {
                err = 0;
                break;
            }
        }
        if (elapsedSecs >= retryTimeoutSecs) {
            break;
        }
        sleep(retryIntervalSecs);
        elapsedSecs += retryIntervalSecs;
    }
    return err;
}

/* compare the contents of the buffers and print out any differences*/
int compare(const char * buffer1, const char * buffer2, int bytes)
{
    assert(buffer1 != NULL && buffer2 != NULL && bytes <= BLOCKSIZE);
    if (!memcmp(buffer1, buffer2, bytes)) {
        return 0;
    }
    fprintf(stderr, "differences:\n");
    for (int i = 0; i < bytes; i++) {
        if (buffer1[i] != buffer2[i]) {
            fprintf(stderr, "0x%x: 0x%.2x 0x%.2x\n", i, (unsigned char)buffer1[i], (unsigned char)buffer2[i]);
        }
    }
    return 1;
}

/* write the progress of the test in percent to stderr */
int progress(int tenpercentdone, int blocksdone, int blocks)
{
    while ((blocksdone * 10)/blocks > tenpercentdone) {
        tenpercentdone++;
        fprintf(stderr, "%d%%\n", tenpercentdone * 10);
    }
    return tenpercentdone;
}

int main(int argc, char ** argv)
{
    fprintf(stderr, "e2e-storage-tester 0.4, args:");
    for (int i = 1; i < argc; i++) {
        fprintf(stderr, " %s", argv[i]);
    }
    fprintf(stderr, "\n");

    const char * device = "";
    bool b_read = false;
    bool b_write = false;
    bool verify = false;
    int blocksdone = 0;
    int blocks = 1;
    int timeoutSecs = 50;
    int timeoutInterval = 5;
    int err = 0;

    for (int i = 1; i < argc; i++)
    {
        if (argv[i][0] == '-') {
            if (strlen(argv[i]) > 1) {
                switch (argv[i][1]) {
                    case 'h': help("","");      break;
                    case 'n':
                        if (i+1 >= argc) {
                            help("no of blocks not specified", "");
                        }
                        i++;
                        blocks = atoi(argv[i]);
                        if (blocks <= 0) {
                            help("invalid block count", argv[i]);
                        }
                    break;
                    case 't':
                        if (i+1 >= argc) {
                            help("timeout not specified", "");
                        }
                        i++;
                        char * endptr;
                        timeoutSecs = strtol(argv[i], &endptr, 10);
                        if (endptr == argv[i] || timeoutSecs < 0) {
                            help("invalid timeout", argv[i]);
                        }
                    break;
                    case 'r': b_read = true;    break;
                    case 'v': verify = true;    break;
                    case 'w': b_write = true;   break;
                    default : help("unrecognized option", argv[i]); break;
                }
            } else {
                help("invalid parameters", argv[i]);
            }
        } else {
            if (device[0] == '\0') {
                device = argv[i];
            } else {
                help("device already specified", argv[i]);
            }
        }
    }
    if (b_write == false && verify == false && b_read == false) {
        help("specify -r, -v or -w", "");
    }

    if (device[0] == '\0') {
        help("device not specified", "");
    }

    int fd;

    if (b_write == true) {
        fd = open(device, O_WRONLY | O_DIRECT);

        if (fd < 0) {
            fprintf(stderr, "could not open %s, error: %s\n", device, strerror(errno));
            exit(-1);
        }
        fprintf(stderr, "writing:\n");
        int tenpercentdone = 0;
        for (blocksdone = 0; blocksdone < blocks; blocksdone++) {
            err = ioAttempt(
                fd,
                pattern_buffer,
                EWRITE,
                blocksdone,
                timeoutInterval,
                timeoutSecs);
            if (err != 0) {
                break;
            }
            tenpercentdone = progress(tenpercentdone, blocksdone, blocks);
        }
        close(fd);
        if (err < 0) {
            fprintf(stderr, "write failure\n");
        }
    }
    if (verify == true || b_read == true) {
        fd = open(device, O_RDONLY | O_DIRECT);

        if (fd < 0) {
            fprintf(stderr, "could not open %s, error: %s\n", device, strerror(errno));
            exit(-1);
        }
        if (verify == true) {
            fprintf(stderr, "verifying:\n");
        } else {
            fprintf(stderr, "reading:\n");
        }
        int tenpercentdone = 0;
        for (blocksdone = 0; blocksdone < blocks; blocksdone++) {
            err = lseek(fd, (int64_t)(blocksdone * BLOCKSIZE), 0);
            if (err < 0) {
                fprintf(stderr, "could not seek to block %d, error: %s\n", blocksdone, strerror(errno));
                break;
            }
            err = ioAttempt(
                fd,
                read_buffer,
                EREAD,
                blocksdone,
                timeoutInterval,
                timeoutSecs);
            if (err < 0) {
                break;
            }
            if (verify == true) {
                blockPattern(blocksdone, pattern_buffer, BLOCKSIZE);
                int res = compare(read_buffer, pattern_buffer, BLOCKSIZE);
                if (res != 0) {
                    fprintf(stderr, "buffers mismatch at block %d\n", blocksdone);
                    err = -1;
                    break;
                }
            }
            tenpercentdone = progress(tenpercentdone, blocksdone, blocks);
        }
        close(fd);

        if (err < 0) {
            fprintf(stderr, "read failure\n");
        }
    }
    fprintf(stdout, "%d", blocksdone);
    return(err);
}
