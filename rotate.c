#include <math.h>
#include <stdint.h>

// Inline fixed-point bilinear interpolation function
static inline void get_bilinear_pixel(
    const uint8_t* input,
    int            width,
    int            height,
    double         x,
    double         y,
    uint8_t*       r,
    uint8_t*       g,
    uint8_t*       b) {
	// Compute integral parts
	int x_floor = (int) floor(x);
	int y_floor = (int) floor(y);

	// Check bounds for bilinear interpolation
	// We need x_floor, y_floor, x_floor+1, y_floor+1 to be valid indices
	if (x_floor < 0 || y_floor < 0 || x_floor >= width - 1 || y_floor >= height - 1) {
		// Out of bounds: return white
		*r = 255;
		*g = 255;
		*b = 255;
		return;
	}

	// Compute fractional parts in fixed-point Q16 (1.0 = 65536)
	// x_frac = fraction(x), y_frac = fraction(y)
	double x_frac_d = x - x_floor;
	double y_frac_d = y - y_floor;

	int32_t x_frac = (int32_t) (x_frac_d * 65536.0);
	int32_t y_frac = (int32_t) (y_frac_d * 65536.0);

	int32_t one_minus_x = 65536 - x_frac;
	int32_t one_minus_y = 65536 - y_frac;

	// Compute weights (Q16)
	// W00 = (1 - x_frac)*(1 - y_frac)
	// W10 = x_frac*(1 - y_frac)
	// W01 = (1 - x_frac)*y_frac
	// W11 = x_frac*y_frac
	// All results fit into 32-bit safely.
	int32_t W00 = (int32_t) (((int64_t) one_minus_x * one_minus_y) >> 16);
	int32_t W10 = (int32_t) (((int64_t) x_frac * one_minus_y) >> 16);
	int32_t W01 = (int32_t) (((int64_t) one_minus_x * y_frac) >> 16);
	int32_t W11 = (int32_t) (((int64_t) x_frac * y_frac) >> 16);

	int stride = width * 3;

	const uint8_t* p00 = input + y_floor * stride + x_floor * 3;
	const uint8_t* p10 = input + y_floor * stride + (x_floor + 1) * 3;
	const uint8_t* p01 = input + (y_floor + 1) * stride + x_floor * 3;
	const uint8_t* p11 = input + (y_floor + 1) * stride + (x_floor + 1) * 3;

	// Interpolate each channel using fixed-point arithmetic.
	// Final = (p00*C00 + p10*C10 + p01*C01 + p11*C11) >> 16, with rounding.
	// We'll add half (32768) before shifting for rounding.
	int32_t R = ((int32_t) p00[0] * W00) + ((int32_t) p10[0] * W10) +
	            ((int32_t) p01[0] * W01) + ((int32_t) p11[0] * W11);

	int32_t G = ((int32_t) p00[1] * W00) + ((int32_t) p10[1] * W10) +
	            ((int32_t) p01[1] * W01) + ((int32_t) p11[1] * W11);

	int32_t B = ((int32_t) p00[2] * W00) + ((int32_t) p10[2] * W10) +
	            ((int32_t) p01[2] * W01) + ((int32_t) p11[2] * W11);

	// Add 0x8000 for rounding and shift right by 16
	*r = (uint8_t) ((R + 32768) >> 16);
	*g = (uint8_t) ((G + 32768) >> 16);
	*b = (uint8_t) ((B + 32768) >> 16);
}

void rotate_image_bilinear(
    const uint8_t* input,
    uint8_t*       output,
    int            input_width,
    int            input_height,
    int            output_width,
    int            output_height,
    double         angle_radians) {
	// Precompute cos and sin of angle
	double cos_angle = cos(angle_radians);
	double sin_angle = sin(angle_radians);

	// Precompute centers
	double cx_input  = (input_width - 1) / 2.0;
	double cy_input  = (input_height - 1) / 2.0;
	double cx_output = (output_width - 1) / 2.0;
	double cy_output = (output_height - 1) / 2.0;

	int output_stride = output_width * 3;

	for (int y = 0; y < output_height; y++) {
		double y_rel = y - cy_output;
		for (int x = 0; x < output_width; x++) {
			double x_rel = x - cx_output;

			// Rotate back to source coordinates
			double src_x = x_rel * cos_angle + y_rel * sin_angle + cx_input;
			double src_y = -x_rel * sin_angle + y_rel * cos_angle + cy_input;

			uint8_t r, g, b;
			get_bilinear_pixel(input, input_width, input_height, src_x, src_y, &r, &g, &b);

			uint8_t* dst = output + y * output_stride + x * 3;
			dst[0]       = r;
			dst[1]       = g;
			dst[2]       = b;
		}
	}
}
