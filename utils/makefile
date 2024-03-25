include settings.mk

# Set compiler and archiver options
CC=gcc
AR=ar
ARFLAGS=rcs

# Set prerrequisites
SRCS_C += $(shell find src/ -iname "*.c")
SRCS_H += $(shell find src/ -iname "*.h")
DEPS = $(foreach STL,$(STATIC_LIBPATHS),$(STL:%=%/bin/lib$(notdir $(STL)).a))

# Set header paths to (-I)nclude
IDIRS += $(addsuffix /src,$(STATIC_LIBPATHS) .)

# Set library paths to (-L)ook
LIBDIRS = $(addsuffix /bin,$(STATIC_LIBPATHS))

# Set intermediate objects
OBJS = $(patsubst src/%.c,obj/%.o,$(SRCS_C))

# Set binary targets
BIN = bin/lib$(shell cd . && pwd | xargs basename).a

.PHONY: all
all: debug

.PHONY: debug
debug: CFLAGS = $(CDEBUG)
debug: $(BIN)

.PHONY: release
release: CFLAGS = $(CRELEASE)
release: $(BIN)

.PHONY: clean
clean:
	-rm -rfv obj bin

$(BIN): $(OBJS) | $(dir $(BIN))
	$(AR) $(ARFLAGS) -o "$@" $^

obj/%.o: src/%.c $(SRCS_H) $(DEPS) | $(dir $(OBJS))
	$(CC) $(CFLAGS) -fPIC -c -o "$@" $< $(IDIRS:%=-I%)

.SECONDEXPANSION:
$(DEPS): $$(shell find $$(patsubst %bin/,%src/,$$(dir $$@)) -iname "*.c" -or -iname "*.h")
	make --no-print-directory -C $(patsubst %bin/,%,$(dir $@)) 3>&1 1>&2 2>&3 | sed -E 's,(src/)[^ ]+\.(c|h)\:,$(patsubst %bin/,%,$(dir $@))&,' 3>&2 2>&1 1>&3

$(sort $(dir $(BIN) $(OBJS))):
	mkdir -pv $@
