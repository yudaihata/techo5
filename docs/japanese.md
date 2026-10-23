# Japanese support in this fork

This fork keeps the official Go module paths and the existing Screen language mechanism. It adds
Japanese without replacing the settings system or changing how Home Assistant's speech pipeline works.

Choose **Settings > General > Screen language > 日本語**, or choose **日本語** in Home Assistant's
**Screen language** select. Switching back to English or Match all is supported. The saved value is
`ja`. This translates the dates, weather, calendar phrases, light/media/climate sheets and update
dialogs that upstream already localizes. The general settings menu remains English, as it does for
upstream's other languages. The assistant's recognition and reply language is set in Home Assistant.
Japanese weather, rain-map, camera and return-to-clock phrases are included, also under Match all.
Long Japanese replies wrap without splitting UTF-8 characters; very long replies can still be
shortened vertically by upstream's existing layout.

## Font

Japanese needs a fallback font once per device. It lives in persistent data, so ordinary firmware
updates preserve it. Download the pinned static Noto Sans CJK JP and its license:

```sh
python3 tools/japanese/fetch-font.py /tmp/techo5-japanese-fonts
```

The downloader verifies SHA-256 before replacing a font. Upload `NotoSansCJKjp-Regular.otf` using
**Fonts of your own** on the setup page's **Screen & Photos** tab, or copy it over SSH to
`/data/misc/techo5/fonts/`. Keep the downloaded SIL Open Font License with redistributed copies.
The same mechanism supplies Japanese glyphs regardless of the selected primary font.

## Branches and official updates

- `main` is an unmodified snapshot of the official repository. The local remote `upstream` points to
  `HuskerMinion/techo5`, and `origin` points to this fork.
- `codex/japanese` is the default branch, based on official Show release **v1.6.1**. It keeps the reply
  wrapping fix, language addition, and maintenance tools in separate commits. Japanese translations
  are mostly in additional `japanese.go` files to reduce conflicts.
- `codex/japanese-reply-wrap-v1.4.0` preserves the first tested reply fix for the older installed build.

**Check upstream Show release** runs daily at 06:17 JST and can also be run manually in Actions.
It merges the newest stable official Show release into a new `codex/upstream-vX.Y.Z` branch, opens a
PR against `codex/japanese`, and explicitly starts **Japanese fork checks**. It never merges the PR
or updates a device automatically. Rewritten upstream history or overlapping edits stop the merge
for manual resolution; no force-push is used. GitHub can disable scheduled workflows on inactive
public repositories; re-enable or run it manually if that happens.

The repository must allow GitHub Actions to create pull requests. The workflow requests only the
permissions needed to push its branch, open the PR and dispatch the checks. Review a green PR and
merge it, then build/install the resulting commit. The same operation can be run from a clean local
checkout with GitHub CLI authentication:

```sh
git switch codex/japanese
git pull --ff-only origin codex/japanese
tools/japanese/sync-upstream.sh
```

To refresh the unmodified `main` snapshot separately:

```sh
gh repo sync yudaihata/techo5 --source HuskerMinion/techo5 --branch main
```

If upstream rewrites history, inspect the changes before explicitly resetting the mirror. Never
reset the Japanese branch to upstream, which would discard the additional commits.

## Checks and device builds

**Japanese fork checks** runs the Linux tests and vet for Show, Dot, Spot and Panel. It also exercises
Japanese selection and switching back, localized dates and update dialogs, voice triggers, and
Japanese reply rendering using the verified CJK font. Successful runs produce an ARMv7 Show daemon
artifact; it is a development artifact, not an official signed OTA release.

For the standard ARM Show image (including the first-generation Echo Show 8/crown), a clean
checkout can prepare a personal rootfs without installing anything:

```sh
python3 tools/japanese/build-show.py
```

It verifies the official release manifest signature and rootfs hash, stamps the exact source commit,
replaces only the daemon and release label, checks all archive members and metadata, and writes the
rootfs and hashes under `bin/japanese/`. This is not the athena or skylight image. Never install an
image intended for a different board.

Follow [the build guide](building.md) and [contributing instructions](../CONTRIBUTING.md) to try a
daemon temporarily before preparing a persistent image. For a persistent image use the matching
official release's verified rootfs and replace only the daemon and release label, keeping the kernel
and the working device slot as rollback. Preserve settings and verify a five-minute automatic slot
commit after reboot. An official OTA update can replace these local changes; rebuilding from this
branch preserves them. The official signing keys and update feed are unchanged.

To keep the fork changes, turn off **Install updates automatically** for the official feed and
install reviewed fork builds instead. The daily upstream PR check does not depend on the device
auto-install switch. Keeping official auto-install on means local changes may be replaced overnight.

## Photo clock (Show and Panel)

Choose **Settings > Display > Clock style > Photo** (or Home Assistant's **Clock style** select).
This keeps a compact date above a regular-weight time, with color weather and temperature beside
it, at the lower left. Clock text stays white over photographs, regardless of the theme's gold photo text.
The photograph keeps its original brightness except for a soft shade behind the lower text. The
selected screen font and 12/24-hour format still apply. Clock position and Date color are not used
by this deliberately fixed photo layout. Other clock styles retain their existing behavior.

Photo works with slideshow Background and Screensaver modes; the Screensaver's Off overlay remains
off. Photo captions stay visible and move the clock above them. On the ordinary home screen the clock
also moves above music, timers and next-up cards. Without a photo it uses the theme's text color.

Photo weather uses the static **Flat** style from [Meteocons](https://github.com/basmilius/meteocons)
(Bas Milius, MIT). Sixteen conditions/day-night variants are vendored with version, package integrity
and source hashes, so the device needs no CDN connection. The 96px transparent PNGs have normalized
outer padding and are scaled/cached for each display size. Regenerate them from the included SVGs
with `tools/japanese/render-weather-icons.py` (CairoSVG2.8.2, Pillow12.0.0, Cairo). Other clock styles
keep their original icons.

## Noto Sans CJK JP as the primary font

Choose **Settings > Display > Font > Noto Sans CJK JP**, or select it in Home Assistant's
**Screen font** entity. Both the full static Regular and Bold cuts are bundled (SIL OFL1.1),
covering Japanese, basic Latin and clock digits in one family, without requiring an uploaded font.
A few extended Latin letters absent from Noto CJK continue to use the existing Go fallback.
The full font data and original license/source hashes are in `echod/internal/lib/fonts/files/notosanscjkjp/`.
They are read through ReaderAt to avoid copying both large font files to the heap. The0.97 size
normalization keeps existing layouts close to their original widths. Other font/style choices remain
available; a skin with its own font still overrides the selected font, as upstream specifies.

## Nest-style home controls (Show / Panel)

Settings > Display > **Nest-style home**, or Home Assistant's `screen_nest_home` switch,
selects an optional home interaction layer. Off preserves upstream gestures. The chosen clock,
photos, font, night schedule and dashboard configuration are kept.

From the home clock:

| Gesture | Result |
| --- | --- |
| Left edge, swipe right | Manual dim night clock; tap or swipe left returns |
| Right edge, swipe left | Existing Home Assistant dashboard |
| Top edge, swipe down | The same dashboard |
| Bottom edge, swipe up | Floating quick-controls panel over the live photo/clock |

The five controls are brightness, volume, **hardware microphone mute**, new alarm, settings.
No feedback button is included. Brightness has ten steps (1–10); tap the sun for auto brightness,
shown by an A. Dragging/tapping the dots sets a manual level and exits auto. Volume is 0–10,
converted to the device's normal media-volume steps. Tap its number/chevron for an independently
adjustable alarm/timer volume row. A tap outside, downward swipe, or 30 seconds without input
closes quick controls. Zero alarm volume is silent; normal device alarm-volume policy applies.
The volume icon returns to the quick-controls menu.

Page changes slide in the direction of the swipe. Quick panels rise from and return to the bottom;
the alarm row rises out from behind media volume. Page snapshots avoid redrawing the dashboard on
every animation frame. Panel backgrounds use a cached alpha mask to preserve the photo underneath.
Opening gestures are consumed through the end of their touchscreen contact; local slider drags
do not turn other swipes into slider input. Calls, alarms, notifications, PIN/setup/update dialogs,
video and voice turns take priority over quick controls. Disabling the layer clears its local state.

Mute uses the existing `mute.Flip` driver and persisted microphone state, with no extra wake-word
setting. On Show hardware whose privacy latch can only be released physically, use the microphone
button on the device to unmute; the panel displays a hint while muted. Manual night display changes
only transient display state/backlight, not scheduled night hours or saved brightness.

The smart-home page uses the dashboard already configured on the device, including its existing
controls and touch forwarding. This change does not recreate Nest's Google service integrations or
replace the dashboard's internal layout. An unconfigured dashboard has a local settings entry.

Regression tests: `TestHub*`, `TestLocalSlider*`, `TestOpeningSwipe*`, `TestContactIdentity*`.
`SHOW_PREVIEW=<folder>` exports synthetic home/control previews at 1280×800 and 960×480. Household
reference images and actual framebuffer captures are kept outside the public repository.
