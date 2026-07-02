export const profile = {
  files: {
    'USER.md': 'Preferences',
    'SOUL.md': 'Persona',
    'AGENTS.md': 'Behavior rules',
    'TOOLS.md': 'Tool conventions',
  },
  fileHints: {
    'SOUL.md': 'Persona, tone, and boundaries (ADR-0010 standard profile)',
    'AGENTS.md': 'In-meeting behavior rules and speaking style',
    'TOOLS.md': 'Tool and environment conventions',
    'USER.md':
      'Principal preferences and background (language, confirmation habits, industry constraints). Moderator serves your long-term settings, not per-meeting topics.',
  },
  list: {
    principalRole: 'Principal',
    participantRole: 'Expert',
    loadingIndex: 'Loading profile index…',
    userConfigured: 'USER.md configured',
    userPending: 'USER.md pending edit',
  },
  form: {
    createTitle: 'Add expert',
    editTitle: 'Edit expert',
    description:
      'ID is the profile directory key; display names must be unique. One bot per platform (Discord, Telegram, etc.).',
    idLabel: 'ID',
    idPlaceholder: 'e.g. analyst',
    idHint: 'Lowercase letter first; only a-z, 0-9, _, -. Directory key; cannot change after create',
    nameLabel: 'Display name',
    namePlaceholder: 'e.g. Data analyst',
    nameHint: 'Shown in meetings; must be unique among experts',
    expertiseLabel: 'Expertise (optional)',
    expertiseHint: 'Optional tag, e.g. research, design',
    imLegend: 'IM bindings (one bot per platform)',
    imHint: 'One bot per platform (Discord, Telegram, etc.). More platforms coming soon',
    discordBotLabel: 'Discord bot',
    discordBotHint: 'Expert speaks via this bot on Discord; leave empty to use the host bot',
    discordNone: 'Unbound (speak via host bot)',
    discordSearchPlaceholder: 'Search bot name or ID…',
    imComingSoon: 'Telegram, Slack, and more coming soon',
    error: {
      idRequired: 'Enter an ID',
      idPattern: 'ID must start with a lowercase letter; only a-z, 0-9, _, -',
      idReserved: 'ID "moderator" is reserved',
      nameRequired: 'Enter a display name',
      idDuplicate: 'ID {id} already exists',
      nameDuplicate: 'Display name duplicates {name}',
    },
  },
  principal: {
    backToList: 'Back to {principal} list',
    description:
      'Server generates USER.md from preferences. Use brief templates for meeting topics.',
    preferences: 'Preferences',
    languageLabel: 'Interface language',
    languageHint:
      'Matches Settings → Service locale; Moderator and Reception use this language with you. Follows system settings; change it in Settings.',
    confirmationLabel: 'Review & confirmation habit',
    confirmationHint:
      'How the Moderator presents items and how you review during the confirmation gate—e.g. numbered lists, short summaries, owners and deadlines.',
    contextLabel: 'Background & constraints',
    contextHint:
      'Industry, team, and project constraints for long-term context—not a single meeting topic.',
    contextPlaceholder: 'e.g. mobile game team; focus on actionable conclusions and launch risk',
    confirmationPlaceholder: 'e.g. review numbered lists; prefer short summaries',
    save: 'Save preferences',
    saveSuccess: 'Preferences saved',
    loadFailed: 'Failed to load profile',
    editProfile: 'Edit profile',
    backToPreview: 'Back to preview',
    persona: {
      switchLabel: 'Profile switch',
      navAriaLabel: 'Principal profile list',
      switchHint:
        'Select a profile on the left to preview; meetings use the profile with the green active indicator',
      activeBadge: 'Active',
      activate: 'Activate profile',
      activating: 'Activating…',
      activated: '"{title}" is now the active profile',
      new: 'New profile',
      newPlaceholder: 'Profile title, e.g. Game design discussions',
      create: 'Create',
      created: 'Profile created',
      switched: 'Switched to "{title}"',
      emptyValue: 'Not set',
      newDialogTitle: 'New profile',
      newDialogDescription:
        'Create a preference profile for another domain. The system assigns an ID; you only need a title.',
      newTitleLabel: 'Profile title',
      newTitleHint: 'A preference profile for another domain; the system assigns an ID',
    },
    preview: {
      description: 'Preview of the active profile. Moderator reads this during meeting preparation.',
    },
    presets: {
      confirmation: {
        numbered: {
          label: 'Numbered review',
          value: 'Review numbered lists item by item before approving',
        },
        brief: {
          label: 'Short summary',
          value: 'Prefer short summaries; focus on risks and open items',
        },
        owners: {
          label: 'Owners & deadlines',
          value: 'Require clear owners and deadlines before confirmation',
        },
      },
      context: {
        mobileGame: {
          label: 'Mobile game',
          value: 'Mobile game team; focus on actionable conclusions, launch risk, and release cadence',
        },
        indie: {
          label: 'Indie dev',
          value: 'Solo/indie builder; small steps, low-cost validation, actionable outcomes',
        },
        b2b: {
          label: 'B2B SaaS',
          value: 'B2B SaaS product group; compliance, customer acceptance, cross-team alignment',
        },
      },
    },
  },
  filesEditor: {
    sectionTitle: 'Profile files',
    notCreated: 'Not created',
    defaultHint: 'Markdown profile',
    createOnSave: '· File will be created on save',
    save: 'Save profile',
    saveSuccess: 'Saved {filename}',
    unsavedPreview: 'Preview has unsaved changes · switch to source to save',
    unsaved: 'Unsaved changes',
    switchConfirm: 'Unsaved changes. Switch file anyway?',
    emptyTitle: 'No Markdown profile',
    loadingMarkdown: 'Loading Markdown profile…',
    participantDescription:
      'Edit SOUL.md / AGENTS.md / TOOLS.md to define Participant persona and in-meeting behavior.',
    participantBack: 'Back to {participant} list',
    participantEmptyHint:
      'Standard files are SOUL.md, AGENTS.md, and TOOLS.md. Missing files are auto-created from templates when you open this page.',
    filesCount: '{present}/{total} files',
    editAria: 'Edit {name}',
    deleteAria: 'Delete {name}',
  },
  state: {
    loadingPrincipal: 'Loading preferences…',
    loadProfileFailed: 'Failed to load profile',
  },
} as const
