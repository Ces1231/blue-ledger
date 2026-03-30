import React, { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';

export interface AchievementNotificationProps {
  id?: string;
  title: string;
  subtitle?: string;
  xpGained?: number;
  badgeName?: string;
  type?: 'badge' | 'level_up' | 'streak' | 'daily_login';
  onDismiss?: () => void;
  autoDismiss?: number; // milliseconds
  icon?: string; // emoji or icon class
}

export const AchievementNotification: React.FC<AchievementNotificationProps> = ({
  id = Math.random().toString(36),
  title,
  subtitle,
  xpGained,
  badgeName,
  type = 'badge',
  onDismiss,
  autoDismiss = 5000,
  icon,
}) => {
  const [isVisible, setIsVisible] = useState(true);

  useEffect(() => {
    if (autoDismiss > 0) {
      const timer = setTimeout(() => {
        setIsVisible(false);
        onDismiss?.();
      }, autoDismiss);

      return () => clearTimeout(timer);
    }
  }, [autoDismiss, onDismiss]);

  const getTypeIcon = () => {
    if (icon) return icon;
    switch (type) {
      case 'badge':
        return '🏆';
      case 'level_up':
        return '📈';
      case 'streak':
        return '🔥';
      case 'daily_login':
        return '📅';
      default:
        return '⭐';
    }
  };

  const getTypeColor = () => {
    switch (type) {
      case 'badge':
        return 'from-yellow-400 to-amber-500';
      case 'level_up':
        return 'from-purple-400 to-pink-500';
      case 'streak':
        return 'from-orange-400 to-red-500';
      case 'daily_login':
        return 'from-blue-400 to-cyan-500';
      default:
        return 'from-green-400 to-emerald-500';
    }
  };

  return (
    <AnimatePresence mode="wait">
      {isVisible && (
        <motion.div
          key={id}
          initial={{ opacity: 0, y: -50, scale: 0.9 }}
          animate={{ opacity: 1, y: 0, scale: 1 }}
          exit={{ opacity: 0, y: -30, scale: 0.9 }}
          transition={{
            type: 'spring',
            stiffness: 300,
            damping: 30,
          }}
          className="fixed top-6 left-1/2 transform -translate-x-1/2 z-50 pointer-events-auto"
        >
          <div className={`relative bg-gradient-to-r ${getTypeColor()} rounded-lg shadow-2xl overflow-hidden`}>
            {/* Animated background shimmer */}
            <motion.div
              className="absolute inset-0 bg-white opacity-20"
              initial={{ x: '-100%' }}
              animate={{ x: '100%' }}
              transition={{ duration: 2, repeat: Infinity }}
            />

            <div className="relative px-6 py-4">
              <div className="flex items-start gap-4">
                {/* Icon with bounce animation */}
                <motion.div
                  className="text-4xl flex-shrink-0"
                  animate={{ scale: [1, 1.1, 1], rotate: [0, 5, -5, 0] }}
                  transition={{ duration: 0.6, delay: 0.2 }}
                >
                  {getTypeIcon()}
                </motion.div>

                <div className="flex-1">
                  {/* Title */}
                  <motion.h3
                    className="text-white font-bold text-lg"
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    transition={{ delay: 0.1 }}
                  >
                    {title}
                  </motion.h3>

                  {/* Subtitle */}
                  {subtitle && (
                    <motion.p
                      className="text-white text-sm opacity-90 mt-1"
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      transition={{ delay: 0.2 }}
                    >
                      {subtitle}
                    </motion.p>
                  )}

                  {/* Badge Name */}
                  {badgeName && (
                    <motion.p
                      className="text-white text-xs opacity-75 mt-2 italic"
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      transition={{ delay: 0.3 }}
                    >
                      Badge: {badgeName}
                    </motion.p>
                  )}

                  {/* XP Gained */}
                  {xpGained !== undefined && xpGained > 0 && (
                    <motion.div
                      className="mt-3 flex items-center gap-2 bg-white bg-opacity-20 rounded px-3 py-1 w-fit"
                      initial={{ opacity: 0, scale: 0.8 }}
                      animate={{ opacity: 1, scale: 1 }}
                      transition={{ delay: 0.4 }}
                    >
                      <span className="text-yellow-200 font-bold">+{xpGained} XP</span>
                    </motion.div>
                  )}
                </div>

                {/* Close button */}
                <motion.button
                  onClick={() => {
                    setIsVisible(false);
                    onDismiss?.();
                  }}
                  className="text-white hover:bg-white hover:bg-opacity-20 rounded-full p-1 transition-colors flex-shrink-0"
                  whileHover={{ scale: 1.1 }}
                  whileTap={{ scale: 0.95 }}
                >
                  <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                    <path
                      fillRule="evenodd"
                      d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
                      clipRule="evenodd"
                    />
                  </svg>
                </motion.button>
              </div>

              {/* Progress bar */}
              {autoDismiss > 0 && (
                <motion.div
                  className="absolute bottom-0 left-0 h-1 bg-white opacity-50"
                  initial={{ scaleX: 1 }}
                  animate={{ scaleX: 0 }}
                  transition={{ duration: autoDismiss / 1000, ease: 'linear' }}
                  style={{ transformOrigin: 'left' }}
                />
              )}
            </div>
          </div>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default AchievementNotification;
