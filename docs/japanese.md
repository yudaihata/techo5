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
