// Package heap measures the C heap, for tests that check what a call into
// Objective-C leaves allocated.
package heap

/*
#include <malloc/malloc.h>

static size_t
heap_in_use(void)
{
    malloc_statistics_t stats;
    malloc_zone_statistics(NULL, &stats);
    return stats.size_in_use;
}
*/
import "C"

// InUse returns the bytes allocated in every malloc zone of the process.
func InUse() int {
	return int(C.heap_in_use())
}
