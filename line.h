#include <stdint.h>

// Define a struct to hold all of the input/output parameters:
typedef struct
{
	int     imgWidth;       // Width of the image in pixels
	int     startX;         // Starting X coordinate of the line
	int     startY;         // Starting Y coordinate of the line
	int     width;          // Length of the line to sample
	int     dx;             // LineSetup.dx
	int     dy;             // LineSetup.dy
	int     stepY;          // LineSetup.stepY
	int32_t gradient;       // LineSetup.gradient (fixed-point 16.16)
	int     whiteThreshold; // Threshold for classifying a pixel as white vs black
} IterateLineSubpixelBakedC_Args;

void IterateLineSubpixelBakedC(IterateLineSubpixelBakedC_Args* args, const uint8_t* pixels, int32_t* outWhite, int32_t* outBlack, int32_t* outTransitions);
