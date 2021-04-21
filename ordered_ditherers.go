package dither

// This file contains matrices I've found from around the Internet. They can
// be used with PixelMapperFromMatrix.

// OrderedDitherMatrix is used to hold a matrix used for ordered dithering. This
// is useful if you find a matrix somewhere and would like to try it out. You can
// create this struct and then give it to PixelMapperFromMatrix.
//
// The matrix must be rectangular - each slice inside the first one must be the same
// length.
//
// Max is the value all the matrix values will be divided by. Usually this is the
// product of the dimensions of the matrix (x*y), or the greatest value in the matrix
// plus one. For diagonal matrices or other matrices with repeated values, it is the
// latter.
//
// Leaving Max as 0 will cause a panic.
//
// Matrix values should almost always range from 0 to Max-1. If the matrix you found
// ranges from 1 to Max, just subtract 1 from every value when programming it.
type OrderedDitherMatrix struct {
	Matrix [][]uint `json:"matrix"`
	Max    uint     `json:"max"`
}

// ClusteredDot4x4 comes from http://caca.zoy.org/study/part2.html
//
// It is not diagonal, so the dots form a grid.
var ClusteredDot4x4 = OrderedDitherMatrix{
	Matrix: [][]uint{
		{12, 5, 6, 13},
		{4, 0, 1, 7},
		{11, 3, 2, 8},
		{15, 10, 9, 14},
	},
	Max: 16,
}

// ClusteredDotDiagonal8x8 comes from http://caca.zoy.org/study/part2.html
//
// They say it "mimics the halftoning techniques used by newspapers". It is called
// "Diagonal" because the resulting dot pattern is at a 45 degree angle.
var ClusteredDotDiagonal8x8 = OrderedDitherMatrix{
	Matrix: [][]uint{
		{24, 10, 12, 26, 35, 47, 49, 37},
		{8, 0, 2, 14, 45, 59, 61, 51},
		{22, 6, 4, 16, 43, 57, 63, 53},
		{30, 20, 18, 28, 33, 41, 55, 39},
		{34, 46, 48, 36, 25, 11, 13, 27},
		{44, 58, 60, 50, 9, 1, 3, 15},
		{42, 56, 62, 52, 23, 7, 5, 17},
		{32, 40, 54, 38, 31, 21, 19, 29},
	},
	Max: 64,
}

// Vertical5x3 comes from http://caca.zoy.org/study/part2.html
//
// They say it "creates artistic vertical line artifacts".
var Vertical5x3 = OrderedDitherMatrix{
	Matrix: [][]uint{
		{9, 3, 0, 6, 12},
		{10, 4, 1, 7, 13},
		{11, 5, 2, 8, 14},
	},
	Max: 15,
}

// Horizontal3x5 is my rotated version of Vertical5x3.
var Horizontal3x5 = OrderedDitherMatrix{
	Matrix: [][]uint{
		{9, 10, 11},
		{3, 4, 5},
		{0, 1, 2},
		{6, 7, 8},
		{12, 13, 14},
	},
	Max: 15,
}

// ClusteredDotDiagonal6x6 comes from Figure 5.4 of the book Digital Halftoning by
// Robert Ulichney. In the book it's called "M = 3". It can represent "19 levels
// of gray". Its dimensions are 6x6, but as a diagonal matrix it is 7x7. It is called
// "Diagonal" because the resulting dot pattern is at a 45 degree angle.
var ClusteredDotDiagonal6x6 = OrderedDitherMatrix{
	Matrix: [][]uint{
		// 1 was subtracted from all original values so the range of values begins
		// at 0. Also the diagonal matrix was converted into a rectangular one.
		{8, 6, 7, 9, 11, 10},
		{5, 0, 1, 12, 17, 16},
		{4, 3, 2, 13, 14, 15},
		{9, 11, 10, 8, 6, 8},
		{12, 17, 16, 5, 0, 1},
		{13, 14, 15, 4, 3, 2},
	},
	Max: 18, // (x*y)/2 because it's diagonal
}

// ClusteredDotDiagonal8x8_2 comes from Figure 5.4 of the book Digital Halftoning by
// Robert Ulichney. In the book it's called "M = 4". It can represent "33 levels
// of gray". Its dimensionsare 8x8, but as a diagonal matrix it is 9x9. It is called
// "Diagonal" because the resulting dot pattern is at a 45 degree angle.
//
// It is almost identical to ClusteredDotDiagonal8x8, but worse because it can
// represent fewer gray levels. There is not much point in using it.
var ClusteredDotDiagonal8x8_2 = OrderedDitherMatrix{
	Matrix: [][]uint{
		// 1 was subtracted from all original values so the range of values begins
		// at 0. Also the diagonal matrix was converted into a rectangular one.
		{13, 11, 12, 15, 18, 20, 19, 16},
		{4, 3, 2, 9, 27, 28, 29, 22},
		{5, 0, 1, 10, 26, 31, 30, 21},
		{8, 6, 7, 14, 23, 25, 24, 17},
		{18, 20, 19, 16, 13, 11, 12, 15},
		{27, 28, 29, 22, 4, 3, 2, 9},
		{26, 31, 30, 21, 5, 0, 1, 10},
		{23, 25, 24, 17, 8, 6, 7, 14},
	},
	Max: 32, // (x*y)/2 because it's diagonal
}

// ClusteredDotDiagonal16x16 comes from Figure 5.4 of the book Digital Halftoning by
// Robert Ulichney. In the book it's called "M = 8". It can represent "129 levels
// of gray". Its dimensions are 16x16, but as a diagonal matrix it is 17x17. It is
// called "Diagonal" because the resulting dot pattern is at a 45 degree angle.
var ClusteredDotDiagonal16x16 = OrderedDitherMatrix{
	Matrix: [][]uint{
		// 1 was subtracted from all original values so the range of values begins
		// at 0.
		{63, 58, 50, 40, 41, 51, 59, 60, 64, 69, 77, 87, 86, 76, 68, 67},
		{57, 33, 27, 18, 19, 28, 34, 52, 70, 94, 100, 109, 108, 99, 93, 75},
		{49, 26, 13, 11, 12, 15, 29, 44, 78, 101, 114, 116, 115, 112, 98, 83},
		{39, 17, 4, 3, 2, 9, 20, 42, 87, 110, 123, 124, 125, 118, 107, 85},
		{38, 16, 5, 0, 1, 10, 21, 43, 89, 111, 122, 127, 126, 117, 106, 84},
		{48, 25, 8, 6, 7, 14, 30, 45, 79, 102, 119, 121, 120, 113, 97, 82},
		{56, 32, 24, 23, 22, 31, 35, 53, 71, 95, 103, 104, 105, 96, 92, 74},
		{62, 55, 47, 37, 36, 46, 54, 61, 65, 72, 80, 90, 91, 81, 73, 66},
		{64, 69, 77, 87, 86, 76, 68, 67, 63, 58, 50, 40, 41, 51, 59, 60},
		{70, 94, 100, 109, 108, 99, 93, 75, 57, 33, 27, 18, 19, 28, 34, 52},
		{78, 101, 114, 116, 115, 112, 98, 83, 49, 26, 13, 11, 12, 15, 29, 44},
		{87, 110, 123, 124, 125, 118, 107, 85, 39, 17, 4, 3, 2, 9, 20, 42},
		{89, 111, 122, 127, 126, 117, 106, 84, 38, 16, 5, 0, 1, 10, 21, 43},
		{79, 102, 119, 121, 120, 113, 97, 82, 48, 25, 8, 6, 7, 14, 30, 45},
		{71, 95, 103, 104, 105, 96, 92, 74, 56, 32, 24, 23, 22, 31, 35, 53},
		{65, 72, 80, 90, 91, 81, 73, 66, 62, 55, 47, 37, 36, 46, 54, 61},
	},
	Max: 128, // (x*y)/2 because it's diagonal
}

// ClusteredDot6x6 comes from Figure 5.9 of the book Digital Halftoning by
// Robert Ulichney. It can represent "37 levels of gray". It is not diagonal.
var ClusteredDot6x6 = OrderedDitherMatrix{
	Matrix: [][]uint{
		// 1 was subtracted from all original values so the range of values begins
		// at 0.
		{34, 29, 17, 21, 30, 35},
		{28, 14, 9, 16, 20, 31},
		{13, 8, 4, 5, 15, 19},
		{12, 3, 0, 1, 10, 18},
		{27, 7, 2, 6, 23, 24},
		{33, 26, 11, 22, 25, 32},
	},
	Max: 36,
}

// ClusteredDotSpiral5x5 comes from Figure 5.13 of the book Digital Halftoning by
// Robert Ulichney. It can represent "26 levels of gray". Its dimensions are 5x5.
//
// Instead of alternating dark and light dots like the other clustered-dot
// matrices, the dark parts grow to fill the area.
var ClusteredDotSpiral5x5 = OrderedDitherMatrix{
	Matrix: [][]uint{
		// 1 was subtracted from all original values so the range of values begins
		// at 0.
		{20, 21, 22, 23, 24},
		{19, 6, 7, 8, 9},
		{18, 5, 0, 1, 10},
		{17, 4, 3, 2, 11},
		{16, 15, 14, 13, 12},
	},
	Max: 25,
}

// ClusteredDotHorizontalLine comes from Figure 5.13 of the book Digital Halftoning by
// Robert Ulichney. It can represent "37 levels of gray". Its dimensions are 6x6.
//
// It "clusters pixels about horizontal lines".
var ClusteredDotHorizontalLine = OrderedDitherMatrix{
	Matrix: [][]uint{
		// 1 was subtracted from all original values so the range of values begins
		// at 0.
		{35, 33, 31, 30, 32, 34},
		{23, 21, 19, 18, 20, 22},
		{11, 9, 7, 6, 8, 10},
		{5, 3, 1, 0, 2, 4},
		{17, 15, 13, 12, 14, 16},
		{29, 27, 25, 24, 26, 28},
	},
	Max: 36,
}

// ClusteredDotVerticalLine is my rotated version of ClusteredDotHorizontalLine.
var ClusteredDotVerticalLine = OrderedDitherMatrix{
	Matrix: [][]uint{
		{35, 23, 11, 5, 17, 29},
		{33, 21, 9, 3, 15, 27},
		{31, 19, 7, 1, 13, 25},
		{30, 18, 6, 0, 12, 24},
		{32, 20, 8, 2, 14, 26},
		{34, 22, 10, 4, 16, 28},
	},
	Max: 36,
}

// ClusteredDot8x8 comes from Figure 1.5 of the book Modern Digital Halftoning,
// Second Edition, by Daniel L. Lau and Gonzalo R. Arce. It is like
// ClusteredDotDiagonal8x8, but is not diagonal. It can represent "65 gray-levels".
var ClusteredDot8x8 = OrderedDitherMatrix{
	Matrix: [][]uint{
		// For some reason the values in the book were in the range of 0-64
		// instead of 0-63. I changed the 64 value to 63, so that pure black
		// didn't end up with occasional white dots.
		{3, 9, 17, 27, 25, 15, 7, 1},
		{11, 29, 38, 46, 44, 36, 23, 5},
		{19, 40, 52, 58, 56, 50, 34, 13},
		{31, 48, 60, 63, 62, 54, 42, 21},
		{30, 47, 59, 63, 61, 53, 41, 20},
		{18, 39, 51, 57, 55, 49, 33, 12},
		{10, 28, 37, 45, 43, 35, 22, 4},
		{2, 8, 16, 26, 24, 14, 6, 0},
	},
	Max: 64,
}

// ClusteredDot6x6_2 comes from https://archive.is/71e9G. On the webpage it is
// called "central white point" while ClusteredDot6x6 is called "clustered dots".
//
// It is nearly identical to ClusteredDot6x6.
var ClusteredDot6x6_2 = OrderedDitherMatrix{
	Matrix: [][]uint{
		{34, 25, 21, 17, 29, 33},
		{30, 13, 9, 5, 12, 24},
		{18, 6, 1, 0, 8, 20},
		{22, 10, 2, 3, 4, 16},
		{26, 14, 7, 11, 15, 28},
		{35, 31, 19, 23, 27, 32},
	},
	Max: 36,
}

// ClusteredDot6x6_3 comes from https://archive.is/71e9G. On the webpage it is
// called "balanced centered point".
//
// It is nearly identical to ClusteredDot6x6.
var ClusteredDot6x6_3 = OrderedDitherMatrix{
	Matrix: [][]uint{
		{30, 22, 16, 21, 33, 35},
		{24, 11, 7, 9, 26, 28},
		{13, 5, 0, 2, 14, 19},
		{15, 3, 1, 4, 12, 18},
		{27, 8, 6, 10, 25, 29},
		{32, 20, 17, 23, 31, 34},
	},
	Max: 36,
}

// ClusteredDotDiagonal8x8_3 comes from https://archive.is/71e9G. On the webpage
// it is called "diagonal ordered matrix with balanced centered points".
//
// It is almost identical to ClusteredDotDiagonal8x8, but worse because it can
// represent fewer gray levels. There is not much point in using it.
//
// It is called "Diagonal" because the resulting dot pattern is at a 45 degree angle.
var ClusteredDotDiagonal8x8_3 = OrderedDitherMatrix{
	Matrix: [][]uint{
		// Yes, values are repeated. This is because the diamond matrix is formed
		// by combining two square matrices in a grid. See the linked page for more.
		{13, 9, 5, 12, 18, 22, 26, 19},
		{6, 1, 0, 8, 25, 30, 31, 23},
		{10, 2, 3, 4, 21, 29, 28, 27},
		{14, 7, 11, 15, 17, 24, 20, 16},
		{18, 22, 26, 19, 13, 9, 5, 12},
		{25, 30, 31, 23, 6, 1, 0, 8},
		{21, 29, 28, 27, 10, 2, 3, 4},
		{17, 24, 20, 16, 14, 7, 11, 15},
	},
	Max: 32,
}
