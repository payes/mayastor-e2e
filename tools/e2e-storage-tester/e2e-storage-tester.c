#define _GNU_SOURCE

#include <assert.h>
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
    "\t -t seconds allowed for retried I/O attempts (default 50, <0 means do not retry)\n"
    "\t -v validate\n"
    "\t -w write\n"
    "\t -n number of blocks (default 1)\n";
    if (param[0] != '\0') {
        fprintf(stderr, "error with parameter %s: ", param);
    }
    if (error_message[0] != '\0') {
        fprintf(stderr, "%s, ", error_message);
    }
    fprintf(stderr, "%s\n", usage_message);
    exit(-1);
}

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
    int expiredSecs = 0;
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
            fprintf(stderr, "could not seek at block %d\n", block_no);
        } else {
            int64_t num_bytes;
            if (op == EWRITE) {
                num_bytes = write(fd, buffer, BLOCKSIZE);
            } else {
                num_bytes = read(fd, buffer, BLOCKSIZE);
            }
            if (num_bytes != BLOCKSIZE) {
                fprintf(stderr, "could not %s at block %d, wanted %d, got %ld\n", action_text, block_no, BLOCKSIZE, num_bytes);
                fprintf(stderr, "attempted seconds %d vs timeout %d\n", expiredSecs, retryTimeoutSecs);
                err = -1;
            } else {
                err = 0;
                break;
            }
        }
        if (expiredSecs >= retryTimeoutSecs) {
            break;
        }
        sleep(retryIntervalSecs);
        expiredSecs += retryIntervalSecs;
    }
    return err;
}

int compare(char * buffer1, char * buffer2, int bytes)
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
                        if (blocks == 0) {
                            help("invalid block count", argv[i]);
                        }
                    break;
                    case 't':
                        if (i+1 >= argc) {
                            help("timeout not specified", "");
                        }
                        i++;
                        timeoutSecs = atoi(argv[i]);
                        if (timeoutSecs == 0) {
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
            fprintf(stderr, "could not open %s\n", device);
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
            fprintf(stderr, "could not open %s\n", device);
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
                fprintf(stderr, "could not seek to block %d\n", blocksdone);
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
