import { useState, useCallback } from 'react';
import { AchievementNotificationProps } from '../components/AchievementNotification';
import { soundEffects } from './soundEffects';

interface Notification extends AchievementNotificationProps {
  id: string;
}

/**
 * Hook for managing achievement notifications and sound effects
 */
export const useAchievementNotifications = () => {
  const [notifications, setNotifications] = useState<Notification[]>([]);

  /**
   * Show a new achievement notification
   */
  const showAchievement = useCallback(
    (
      title: string,
      options?: {
        subtitle?: string;
        xpGained?: number;
        badgeName?: string;
        type?: 'badge' | 'level_up' | 'streak' | 'daily_login';
        icon?: string;
        autoDismiss?: number;
      }
    ) => {
      const id = Math.random().toString(36);
      const notif: Notification = {
        id,
        title,
        ...options,
        autoDismiss: options?.autoDismiss ?? 5000,
      };

      // Play appropriate sound
      const soundType = options?.type || 'badge';
      switch (soundType) {
        case 'level_up':
          soundEffects.playLevelUp();
          break;
        case 'streak':
          soundEffects.playStreak();
          break;
        case 'daily_login':
          soundEffects.playSuccess();
          break;
        default:
          soundEffects.playAchievement();
      }

      // Add to queue
      setNotifications((prev) => [...prev, notif]);

      // Auto-remove after dismissal
      if (notif.autoDismiss && notif.autoDismiss > 0) {
        setTimeout(() => {
          removeNotification(id);
        }, notif.autoDismiss + 500); // Add buffer for animation
      }

      return id;
    },
    []
  );

  /**
   * Remove a notification by ID
   */
  const removeNotification = useCallback((id: string) => {
    setNotifications((prev) => prev.filter((n) => n.id !== id));
  }, []);

  /**
   * Clear all notifications
   */
  const clearAll = useCallback(() => {
    setNotifications([]);
  }, []);

  /**
   * Badge unlocked
   */
  const showBadgeUnlocked = useCallback(
    (badgeName: string, xpGained: number = 100) => {
      return showAchievement(`🏆 Badge Unlocked!`, {
        badgeName,
        subtitle: 'New achievement unlocked',
        xpGained,
        type: 'badge',
        icon: '🏆',
      });
    },
    [showAchievement]
  );

  /**
   * Level up
   */
  const showLevelUp = useCallback(
    (newLevel: number, tierName: string = '', xpGained: number = 250) => {
      return showAchievement(`📈 Level Up!`, {
        subtitle: `You've reached level ${newLevel}${tierName ? ` - ${tierName}` : ''}`,
        xpGained,
        type: 'level_up',
        icon: '📈',
      });
    },
    [showAchievement]
  );

  /**
   * Streak milestone
   */
  const showStreakMilestone = useCallback(
    (days: number, xpGained: number = 50) => {
      return showAchievement(`🔥 Streak Milestone!`, {
        subtitle: `${days} days on fire!`,
        xpGained,
        type: 'streak',
        icon: '🔥',
      });
    },
    [showAchievement]
  );

  /**
   * Daily login bonus
   */
  const showDailyLoginBonus = useCallback(
    (xpGained: number = 25, streak: number = 1) => {
      let subtitle = `Day ${streak} login streak!`;
      let bonus = 0;

      if (streak > 0 && streak % 7 === 0) {
        subtitle = `🎉 Weekly milestone! Day ${streak} - Double bonus!`;
        bonus = 50;
      }

      return showAchievement(`📅 Daily Login Bonus!`, {
        subtitle,
        xpGained: xpGained + bonus,
        type: 'daily_login',
        icon: '📅',
      });
    },
    [showAchievement]
  );

  /**
   * XP gained notification (more subtle)
   */
  const showXPGained = useCallback(
    (xp: number, reason: string = '') => {
      return showAchievement(`⭐ +${xp} XP`, {
        subtitle: reason || 'Activity completed',
        type: 'badge',
        icon: '⭐',
        autoDismiss: 3000,
      });
    },
    [showAchievement]
  );

  /**
   * Error notification
   */
  const showError = useCallback(
    (title: string, subtitle?: string) => {
      soundEffects.playError();
      return showAchievement(`❌ ${title}`, {
        subtitle,
        type: 'badge',
        icon: '❌',
        autoDismiss: 4000,
      });
    },
    [showAchievement]
  );

  return {
    notifications,
    showAchievement,
    showBadgeUnlocked,
    showLevelUp,
    showStreakMilestone,
    showDailyLoginBonus,
    showXPGained,
    showError,
    removeNotification,
    clearAll,
  };
};

export default useAchievementNotifications;
