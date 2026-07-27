# Reasonix desktop pet packs

Reasonix ships with an animated Akita pet. Additional Codex-style pet packs can
be installed without rebuilding the app.

On Windows, create this directory:

```text
%AppData%\Reasonix\pets\<pet-id>\
```

Each pack contains:

```text
<pet-id>/
├── pet.json
└── spritesheet.webp
```

Example manifest:

```json
{
  "id": "my-pet",
  "displayName": "My Pet",
  "description": "A custom animated task companion.",
  "spritesheetPath": "spritesheet.webp"
}
```

The folder name and `id` must match. IDs may contain lowercase ASCII letters,
digits, hyphens, and underscores.

The spritesheet uses the Codex/OpenPets 8-column by 9-row layout. Each cell is
192 by 208 pixels. Rows are:

1. idle
2. running right
3. running left
4. waving / approval
5. jumping / success
6. failed
7. waiting
8. running / working
9. review / thinking

After copying the pack, reopen **Settings → Appearance → Pets** and select it.
