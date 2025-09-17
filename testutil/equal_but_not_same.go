package testutil

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

// EqualButNotSame asserts that expected and actual objects are not the same.
// It recursively checks all fields to ensure that no pointers are shared.
// If a pointer, slice or map are nil in either object, the test fails.
func EqualButNotSame(t *testing.T, expected, actual any) {
	t.Helper()

	expectedVal := reflect.ValueOf(expected)
	actualVal := reflect.ValueOf(actual)

	deepEqual(t, expectedVal, actualVal, "")
}

// deepEqual recursively verifies that all values are equal but pointers are different
// except for the Expires field which is explicitly allowed to be shared
func deepEqual(t *testing.T, left, right reflect.Value, fieldPath string) {
	require.True(t, left.IsValid() && right.IsValid(), "invalid value at %s", fieldPath)
	require.Equal(t, left.Type(), right.Type(), "types are not equal at %s", fieldPath)

	switch left.Kind() {
	case reflect.Ptr:
		// Pointers should not be nil and must point to different objects
		require.False(t, left.IsNil(), "nil pointer at %s (left)", fieldPath)
		require.False(t, right.IsNil(), "nil pointer at %s (right)", fieldPath)
		require.NotSame(t, left.Interface(), right.Interface(), "shared pointer at %s", fieldPath)

		deepEqual(t, left.Elem(), right.Elem(), fieldPath)

	case reflect.Slice:
		// Slices should contain some elements and must not share the same underlying array
		require.Equal(t, left.Len(), right.Len(), "slice length mismatch at %s", fieldPath)
		require.NotEmpty(t, left.Len(), "slice must not be empty %s (left)", fieldPath)
		require.NotEmpty(t, right.Len(), "slice must not be empty %s (right)", fieldPath)
		require.NotEqual(t, left.Pointer(), right.Pointer(), "shared slices at %s", fieldPath)

		// Recursively verify slice elements
		for i := 0; i < left.Len(); i++ {
			elemPath := buildPath(fieldPath, "[", anyToString(i), "]")
			deepEqual(t, left.Index(i), right.Index(i), elemPath)
		}

	case reflect.Map:
		// Maps should contain some elements and must not share the same underlying map
		require.Equal(t, left.Len(), right.Len(), "map length mismatch at %s", fieldPath)
		require.NotEmpty(t, left.Len(), "map must not be empty %s (left)", fieldPath)
		require.NotEmpty(t, right.Len(), "map must not be empty %s (right)", fieldPath)
		require.NotEqual(t, left.Pointer(), right.Pointer(), "shared maps at %s", fieldPath)

		// Recursively verify map values
		for _, key := range left.MapKeys() {
			keyStr := anyToString(key.Interface())
			keyPath := buildPath(fieldPath, "[", keyStr, "]")
			originalMapVal := left.MapIndex(key)
			clonedMapVal := right.MapIndex(key)
			deepEqual(t, originalMapVal, clonedMapVal, keyPath)
		}

	case reflect.Struct:
		require.Equal(t, left.Interface(), right.Interface(), "structs are not equal at %s", fieldPath)

		// Fallback to recursive field-by-field comparison
		for i := 0; i < left.NumField(); i++ {
			field := left.Type().Field(i)
			if !field.IsExported() {
				continue // Skip unexported fields
			}

			nestedPath := buildPath(fieldPath, ".", field.Name, "")
			originalFieldVal := left.Field(i)
			clonedFieldVal := right.Field(i)
			deepEqual(t, originalFieldVal, clonedFieldVal, nestedPath)
		}

	default:
		// For primitive types, just verify equality
		require.Equal(t, left.Interface(), right.Interface(), "values not equal at %s", fieldPath)
	}
}

// buildPath builds a field path for error messages
func buildPath(basePath, separator, element, suffix string) string {
	if basePath == "" {
		return element + suffix
	}
	return basePath + separator + element + suffix
}

// anyToString converts an any to a string for path building
func anyToString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	default:
		return reflect.ValueOf(val).String()
	}
}

// ........-------------------============++===============+=+==+++++++++++++++++++
// .........------------------==-=====#%%%%%%%%%#+============++==+==++++++++++++++
// .........--------------------===++#%##%%%##%#####===========+==+=++=++=+++++=+++
//  ..........-----------------=-=+#%#####+++=-=-++#%#===========+==+==++=+==++++=+
//  ..........-----------------==###+#####++=---.. .=++==============+===+=+=+=++=+
//  .........-.---------------=--+=-=####++==---..    -=================++=+==++=++
// .. ........-.--------------==+.  -#####+==---.      ====================++==+++=
// .. ......---.-----------==-==-   .+#++++=----.      =====================++==++=
//  ........--.-------------===-.   -#%%%#+-=-=---.    -=====================++==++
//  ...........------------==++%=  .+#+=+%#+#.=-.--    =====================+==++=+
// .........--------------===###   ++=#%=%+#+ -+#+=.  .--====================+==+==
// ........----------------===%#   +.+###%+#- =+##.=.   -===================++=++++
// ........---------------====##   ++#%%##+#=  -=-     .==================++++++=++
// .......----------------===--+  .+#%#++++#=   ..     =========+=====+++=++=++++++
// .....----------------=-====--  -+==+%%#%#+- =       ============+==++++=++==++=+
// ..----------------==========-  -- #%%#%#+=. =-     ========+===+=++=++++++++++++
// .----------------==-=========. ==-%#++#=-.   =.   ====+=====+==+==+++==++=++++++
// .--------------=--============.=-#%@%%%#+++--=   ===+==++==++=+=+==+++++++++++++
// -------------=-================-=#####%#+-  .-   ==++++++=++=+++++++++++++++++++
// -------------=================- -.+%##+-.+-. .    .+=++=++++++=+++++++++++++++++
// ---------==-================= = --.+%%#=.           =+++++++++++++++++++++++++++
// -------=---===============.  #= =+-==#++++.     .      -++++++++++++++++++++++++
// -----=-==============-  ..  +#=  =%###++       .-          .=+++++++++++++++++++
// --==-===========-  .-==..  -###   --++.        +.       .      .=+++++++++++++++
// ============-  .-=+##=++-. =#%#  =+%##+       #=.                    =++++++++++
// ========== = =##+###=+===- +%%#..%+%+-+#     #+-                       -++++++++
// ========= +==##%%##++-#++- ##=++=.+###%%%   +++.   .. .                  +++++++
// ========-.%++%%%%#+=+-## ++==++++=-..# -=  =+#-   -...        .          ++##+##
// ======++ -%+#%%%%#+=#=#  =+++===+++. . +# -+%#-  .---..........          ++##+#+
// ======+. -%=#%%%%#=+#+. =%%#+++##++=. .=#+=#%+. ..--..-........     -   -++++##+
// +++++++-..%=#%%%%#=+#..###%%#+=+++=---+.=###%=. ..--.--...--...    -.    ++++#++
// ++=++++.%#+=#%%%#+++ =++%%@%%###%##++=.-###%+-..-------.------... .=.     ++##+#
// +++++++.#%%=#%%%#+ .##%%%%%%%###%@@%##%=#%%#=-.--=-=---------...  +=.     =#####
// ++++++ .%%%#+#### ==##%%%@%%#=%@@@%%###+#%%+=--==-===--------....-#=.      +####
// +++++= +%%%%-#- +#+#%%%@%%%#+%@@@@%%%%%%%%+==========--------..--#+--..... .####
// ++++= -.#%%%- #++##%%%%%%%#=%%@@@%#%%%%%%#+==+============------=#=--.. .... ###
// +++++ .#+%% .=%### +%%#%%@%=#%@@@%%%@@%%%+++++++================%+==--..  . . ##
// ++++  .==. =+##%%#%%%%%%%@%=#%@@@@%%%%@@%++++++++=++===========#%++=----.     +#
// +++ .+#..=###%##%#@%#%#%%%%=##@@@%#=%%@%#######++++++++++=+++++%%#++=---..-   =#
// +==--=.+##%%%%+=++++++#@%%%+#%@@@@%+=+#%###%#####++++++++=++++#%%%#++=---..--.-#
// +%%%-+=%#@@%%%+=#%+++#%@%%%##%%@@@@%%#=++#%%%%######+++++=++++%%%%##++===-.  . +
// .- --%%@#@@@%%+++#%##@@@@%%%%%%%######%%##+#%######++#+++++++#%%%%##++++###+=  -
// ..=++%@@%@@@%%%##%#%@@@@@%%%@@@@@%%%#++++++++%###+++###++++++@@%%%%##+==--==++#=
// -+#%%%@@%@@@@%%%%%@@@@@@@@%@@%%%%%%%###++++++++#####+++++++++%@%%%%######+=--..-
// +##%%%%@@@@@@@%%@@@@@@@@@@@@@@@@%%%##++#%#+++++++++++#####+++%@@@%%%#++=====++=-
// #%#@%@%@@@@@%%@@@@@@@@@@@@@%%%%%%%@%%#%%#####++++#+====++++++%@@@@@@@@@%#+==----
// ##%%@%%@@@@%@@@@@@@@@@@@@@@@@%%%%%%%%%%%#%%%%%###%+++======+++##+=+#+++##%%#%%%+
// +#%%@@@%@@@%@@%+#@@@@@@@@@@@@%%%%@@%%%%%@%%%%%%%%@#+++++===-===#++==+++=+++#%#%+
// #%%@@@%%@%%%%@#+#%@@@@%%%@@@@@@@%%%%@@%%%%@%%%%%%@#######+#+==+++=++=======+=++%
// ##@@@@@%%%%@#++++#%%%%%%%%@@@@@@@@@@@@@@@@@@@%%@@@##########+=++++++#++++++###+#
// ++#%@%%%%%#++++#+-###%%%%%%%@@%@@@@@@@@@@@@@@@@@@@##+##+####%#########+#########
// +++++++#++++##+#+ ++####%%%%%%@%@@@@@@@@@@@@@@@%%%####%%####%%#####%#%%#####%%##
// +#++#++++++++#+++ -+#+##%%%%@@%%@@@@@@@@@@@@@@@@@%####%%%%%%%%%%%%##%@@%%%%#%%%%
// +#+++#+#####++++= .=++##%#%%%%%%%@@@@@@@@@@@@@@@%####%%%%@%%@%%%%%%%%%@@%@%%%%%#
// ####+#+++++###+#-..=++##%%%%%%%#%@@@@@@@@@@@@@@@%####%%%%%%%%@%#################
// #++#++##+#++#++#..-=+###%%%%%%#+%@@@@@@@@@@@@@@@@@@@@@@@@@@@@@%#################
// ++#+#++#+##++##=---=++####%%%###@@@%@@@@@@%%@@@@@@@@@@@@@@@@@%%+################
// +##+##++#++#++#--=+=++###%#%####@@@@@@@@@@@@@@@%@@@%%@%%%%%%###=+###############
// +++++##+#++#+++.-===+#####%%####@@@@@@@@%%@@%%@@%%%%%%########+==###############
// +++#++++++#+++..--==+##########%%%@@@@@%%%%%@@%%%%########%##+#=-###############
// +#++#++#+#++++..--===%#########%%%%%%%@%%%%%###%##############+==+##############
// ++++##+++++#++.==--=+@%##+###+#%%%%%%%%%%%####%##%%#%%%#######++==##############
