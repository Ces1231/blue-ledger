import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    './index.html',
    './src/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        // ── Blue Ledger Design Tokens (ported from prototype CSS variables) ──
        navy:    '#001A4D',
        navy2:   '#00245C',
        navy3:   '#002E7A',
        gold:    '#C9A84C',
        gold2:   '#D4B86A',
        gold3:   '#E8D49E',
        'gold-pale': '#FDF8EE',
        'gold-bg':   '#F5E6C8',
        cream:   '#FAF8F3',
        cream2:  '#F2EDE4',
        ink:     '#0D0D0D',
        ink2:    '#2A2A2A',
        muted:   '#6B6657',
        faint:   '#A8A099',
        border:  '#E2DDD4',
        border2: '#CDC7BD',
        success: '#1A6B3A',
        'success-bg':     '#EAF5EE',
        'success-border': '#A8D5BC',
        info:    '#003087',
        'info-bg':     '#EAF0FB',
        'info-border': '#BDD0F5',
        warn:    '#7A4A00',
        'warn-bg':     '#FFF3DC',
        'warn-border': '#F5D98A',
        danger:  '#8B1A1A',
        'danger-bg':     '#FCE4EC',
        'danger-border': '#F4C0D1',
      },
      fontFamily: {
        sans:  ['"Instrument Sans"', 'sans-serif'],
        serif: ['"DM Serif Display"', 'serif'],
        mono:  ['"DM Mono"', 'monospace'],
      },
      borderRadius: {
        DEFAULT: '10px',
        lg:      '14px',
        xl:      '20px',
      },
      boxShadow: {
        DEFAULT: '0 1px 3px rgba(0,0,0,.06), 0 4px 16px rgba(0,0,0,.04)',
        md:      '0 2px 8px rgba(0,0,0,.08), 0 8px 32px rgba(0,0,0,.06)',
      },
    },
  },
  plugins: [],
}

export default config
