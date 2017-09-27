/* 
 * Copyright (c) 2017, SingularityWare, LLC. All rights reserved.
 *
 * Copyright (c) 2015-2017, Gregory M. Kurtzer. All rights reserved.
 * 
 * Copyright (c) 2016-2017, The Regents of the University of California,
 * through Lawrence Berkeley National Laboratory (subject to receipt of any
 * required approvals from the U.S. Dept. of Energy).  All rights reserved.
 * 
 * This software is licensed under a customized 3-clause BSD license.  Please
 * consult LICENSE file distributed with the sources of this project regarding
 * your rights to use or distribute this software.
 *
 * NOTICE.  This Software was developed under funding from the U.S. Department of
 * Energy and the U.S. Government consequently retains certain rights. As such,
 * the U.S. Government has been granted for itself and others acting on its
 * behalf a paid-up, nonexclusive, irrevocable, worldwide license in the Software
 * to reproduce, distribute copies to the public, prepare derivative works, and
 * perform publicly and display publicly, and to permit other to do so. 
 * 
 */

#include <stdio.h>
#include <stdlib.h>
#include <errno.h> 
#include <string.h>
#include <limits.h>
#include <link.h>

#include "config.h"
#include "util/util.h"
#include "util/file.h"
#include "util/registry.h"
#include "util/config_parser.h"
#include "util/message.h"

#ifndef SYSCONFDIR
#error SYSCONFDIR not defined
#endif

#ifndef LIBEXECDIR
#error LIBEXECDIR not defined
#endif


int is_enabled = -1;

int singularity_suid_init(char **envp) {
    ElfW(auxv_t) *auxv;
    char *progname = NULL;

    /* use auxiliary vectors to determine if running privileged */
    while(*envp++ != NULL) {}

    auxv = (ElfW(auxv_t) *)envp;

    for (; auxv->a_type != AT_NULL; auxv++) {
        if ( auxv->a_type == AT_SECURE ) {
            is_enabled = (int)auxv->a_un.a_val;
        }
        if ( auxv->a_type == AT_EXECFN ) {
            progname = (char *)auxv->a_un.a_val;
        }
    }

    if ( is_enabled < 0 ) {
        singularity_message(ERROR, "Failed to determine if program run with SUID or not\n");
        ABORT(255);
    }

    if ( progname == NULL ) {
        singularity_message(ERROR, "Failed to retrieve program name\n");
        ABORT(255);
    }

#ifdef SINGULARITY_SUID

    singularity_message(VERBOSE2, "Running SUID program workflow\n");

    singularity_message(VERBOSE2, "Checking program has appropriate permissions\n");
    if ( is_enabled == 0 ) {
        singularity_message(ERROR, "Installation error, run the following commands as root to fix:\n");
        singularity_message(ERROR, "    sudo chown root:root %s\n", progname);
        singularity_message(ERROR, "    sudo chmod 4755 %s\n", progname);
        if ( getuid() == 0 ) {
            singularity_message(INFO, "\n");
        } else {
            ABORT(255);
        }
    }

    singularity_message(VERBOSE2, "Checking if singularity.conf allows us to run as suid\n");
    if ( ( singularity_config_get_bool(ALLOW_SETUID) <= 0  ) || ( singularity_registry_get("NOSUID") != NULL ) ) {
        return(-1);
    }

#else
    singularity_message(VERBOSE, "Running NON-SUID program workflow\n");

    singularity_message(DEBUG, "Checking program has appropriate permissions\n");
    if ( is_suid(progname) == 0 ) {
        singularity_message(ERROR, "This program must **NOT** be SUID\n");
        ABORT(255);
    }
#endif /* SINGULARITY_SUID */

    return(0);
}

int singularity_suid_enabled(void) {
    return(is_enabled);
}
