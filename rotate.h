#include <stdint.h>

void rotate_image_bilinear(const uint8_t* input,
                           uint8_t*       output,
                           int            input_width,
                           int            input_height,
                           int            output_width,
                           int            output_height,
                           double         angle_radians);
