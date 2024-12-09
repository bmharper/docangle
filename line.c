#include "line.h"

void IterateLineSubpixelBakedC(IterateLineSubpixelBakedC_Args* args, const uint8_t* pixels, int32_t* outWhite, int32_t* outBlack) {
	int     imgWidth       = args->imgWidth;
	int     startX         = args->startX;
	int     startY         = args->startY;
	int     width          = args->width;
	int     dx             = args->dx;
	int     dy             = args->dy;
	int     stepY          = args->stepY;
	int32_t gradient       = args->gradient;
	int     whiteThreshold = args->whiteThreshold;

	int err  = dx - dy;
	int x    = startX;
	int y    = startY;
	int endX = startX + width;
	int line = y * imgWidth;

	int     nWhite = 0;
	int     nBlack = 0;
	int32_t wt     = (int32_t) whiteThreshold;
	int32_t blend  = 0;

	while (x < endX) {
		int32_t va = (int32_t) pixels[line + x];

		// Compute the vertical pixel based on blend
		int32_t wb = blend;
		int32_t vb;
		if (wb < 0) {
			wb = -wb;
			vb = (int32_t) pixels[line + x - imgWidth];
		} else {
			vb = (int32_t) pixels[line + x + imgWidth];
		}

		int32_t diff    = vb - va;
		int32_t blended = va + ((diff * wb) >> 16);

		if (blended > wt) {
			nWhite++;
		} else {
			nBlack++;
		}

		// Update error and coordinates
		int e2 = 2 * err;
		if (e2 > -dy) {
			err -= dy;
			x += 1;
		}
		if (e2 < dx) {
			err += dx;
			y += stepY;
			line += stepY * imgWidth;
			blend -= (int32_t) (stepY << 16);
		}
		blend += gradient;
	}

	// Store the results
	*outWhite = nWhite;
	*outBlack = nBlack;
}
