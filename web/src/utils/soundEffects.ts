/**
 * Sound Effects Manager
 * Centralized management of all game sounds for Blue Ledger
 */

type SoundType = 'success' | 'error' | 'levelup' | 'achievement' | 'streak' | 'click';

interface SoundConfig {
  volume: number;
  muted: boolean;
}

class SoundEffectsManager {
  private audioContext: AudioContext | null = null;
  private sounds: Map<SoundType, HTMLAudioElement> = new Map();
  private config: SoundConfig = {
    volume: 0.7,
    muted: false,
  };

  constructor() {
    this.initAudioContext();
    this.loadSounds();
    this.loadUserPreferences();
  }

  private initAudioContext() {
    if (typeof window !== 'undefined' && window.AudioContext) {
      try {
        this.audioContext = new window.AudioContext();
      } catch (e) {
        console.warn('AudioContext not available:', e);
      }
    }
  }

  private loadSounds() {
    // Using Web Audio API to generate sounds (no external files needed)
    // Or load from CDN/public folder if files exist

    // For now, we'll create audio elements that can be populated later
    const soundTypes: SoundType[] = ['success', 'error', 'levelup', 'achievement', 'streak', 'click'];

    soundTypes.forEach((type) => {
      const audio = new Audio();
      audio.volume = this.config.volume;
      // These URLs can be replaced with actual sound files
      audio.src = this.getSoundURL(type);
      this.sounds.set(type, audio);
    });
  }

  private getSoundURL(type: SoundType): string {
    // Map sound types to URLs
    // You can use Web Audio API or external URLs
    const soundMap: Record<SoundType, string> = {
      success: 'data:audio/wav;base64,UklGRiYAAABXQVZFZm10IBAAAAABAAEAQB8AAAB9AAACABAAZGF0YQIAAAAAAA==',
      error: 'data:audio/wav;base64,UklGRiYAAABXQVZFZm10IBAAAAABAAEAQB8AAAB9AAACABAAZGF0YQIAAAAAAA==',
      levelup: 'data:audio/wav;base64,UklGRiYAAABXQVZFZm10IBAAAAABAAEAQB8AAAB9AAACABAAZGF0YQIAAAAAAA==',
      achievement: 'data:audio/wav;base64,UklGRiYAAABXQVZFZm10IBAAAAABAAEAQB8AAAB9AAACABAAZGF0YQIAAAAAAA==',
      streak: 'data:audio/wav;base64,UklGRiYAAABXQVZFZm10IBAAAAABAAEAQB8AAAB9AAACABAAZGF0YQIAAAAAAA==',
      click: 'data:audio/wav;base64,UklGRiYAAABXQVZFZm10IBAAAAABAAEAQB8AAAB9AAACABAAZGF0YQIAAAAAAA==',
    };
    return soundMap[type];
  }

  /**
   * Play a sound effect
   */
  play(type: SoundType): void {
    if (this.config.muted) return;

    const audio = this.sounds.get(type);
    if (audio) {
      audio.currentTime = 0;
      audio.volume = this.config.volume;
      audio.play().catch((e) => {
        console.warn(`Failed to play sound ${type}:`, e);
      });
    }
  }

  /**
   * Play success sound (XP gained, item purchased)
   */
  playSuccess(): void {
    this.play('success');
  }

  /**
   * Play error sound (failed action, invalid operation)
   */
  playError(): void {
    this.play('error');
  }

  /**
   * Play level up fanfare
   */
  playLevelUp(): void {
    this.play('levelup');
  }

  /**
   * Play achievement unlocked sound
   */
  playAchievement(): void {
    this.play('achievement');
  }

  /**
   * Play streak milestone sound
   */
  playStreak(): void {
    this.play('streak');
  }

  /**
   * Play click sound (UI interaction)
   */
  playClick(): void {
    this.play('click');
  }

  /**
   * Set master volume (0-1)
   */
  setVolume(volume: number): void {
    this.config.volume = Math.max(0, Math.min(1, volume));
    this.sounds.forEach((audio) => {
      audio.volume = this.config.volume;
    });
    this.saveUserPreferences();
  }

  /**
   * Get current volume
   */
  getVolume(): number {
    return this.config.volume;
  }

  /**
   * Mute all sounds
   */
  mute(): void {
    this.config.muted = true;
    this.saveUserPreferences();
  }

  /**
   * Unmute all sounds
   */
  unmute(): void {
    this.config.muted = false;
    this.saveUserPreferences();
  }

  /**
   * Check if sounds are muted
   */
  isMuted(): boolean {
    return this.config.muted;
  }

  /**
   * Preload sounds for faster playback
   */
  preload(): void {
    this.sounds.forEach((audio) => {
      audio.load();
    });
  }

  /**
   * Stop all playing sounds
   */
  stopAll(): void {
    this.sounds.forEach((audio) => {
      audio.pause();
      audio.currentTime = 0;
    });
  }

  /**
   * Save user preferences to localStorage
   */
  private saveUserPreferences(): void {
    localStorage.setItem(
      'soundEffectsConfig',
      JSON.stringify({
        volume: this.config.volume,
        muted: this.config.muted,
      })
    );
  }

  /**
   * Load user preferences from localStorage
   */
  private loadUserPreferences(): void {
    const saved = localStorage.getItem('soundEffectsConfig');
    if (saved) {
      const parsed = JSON.parse(saved);
      this.config = { ...this.config, ...parsed };
      this.sounds.forEach((audio) => {
        audio.volume = this.config.volume;
      });
    }
  }
}

// Export singleton instance
export const soundEffects = new SoundEffectsManager();

export default SoundEffectsManager;
