import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import axios from 'axios';

interface StreakData {
  member_id: string;
  current_streak: number;
  longest_streak: number;
  updated_at: string;
}

interface StreakWidgetProps {
  memberId: string;
  className?: string;
  showAnimation?: boolean;
}

export const StreakWidget: React.FC<StreakWidgetProps> = ({
  memberId,
  className = '',
  showAnimation = true,
}) => {
  const [streak, setStreak] = useState<StreakData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchStreakData();
  }, [memberId]);

  const fetchStreakData = async () => {
    try {
      setLoading(true);
      const response = await axios.get(`/v1/members/${memberId}/streak`);
      setStreak(response.data);
      setError(null);
    } catch (err) {
      console.error('Failed to fetch streak data:', err);
      setError('Failed to load streak data');
    } finally {
      setLoading(false);
    }
  };

  const getStreakColor = (days: number) => {
    if (days >= 30) return 'from-red-500 to-orange-500';
    if (days >= 14) return 'from-orange-500 to-yellow-500';
    if (days >= 7) return 'from-yellow-500 to-amber-500';
    return 'from-gray-400 to-gray-500';
  };

  const getStreakBonus = (days: number) => {
    if (days >= 30) return 250;
    if (days >= 14) return 100;
    if (days >= 7) return 50;
    return 0;
  };

  if (loading) {
    return (
      <div className={`${className}`}>
        <div className="animate-pulse bg-gray-200 rounded-lg h-24" />
      </div>
    );
  }

  if (error || !streak) {
    return null;
  }

  const bonus = getStreakBonus(streak.current_streak);
  const nextMilestone = streak.current_streak < 7 ? 7 : streak.current_streak < 14 ? 14 : 30;

  return (
    <motion.div
      className={`${className}`}
      initial={showAnimation ? { opacity: 0, y: 10 } : {}}
      animate={showAnimation ? { opacity: 1, y: 0 } : {}}
      transition={{ duration: 0.3 }}
    >
      <div className={`bg-gradient-to-br ${getStreakColor(streak.current_streak)} rounded-lg p-4 text-white shadow-lg`}>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            {/* Flame Icon with animation */}
            <motion.div
              className="text-4xl"
              animate={showAnimation ? { scale: [1, 1.1, 1], rotate: [0, 5, -5, 0] } : {}}
              transition={{ duration: 2, repeat: Infinity }}
            >
              🔥
            </motion.div>

            <div>
              <div className="text-sm font-semibold opacity-90">Current Streak</div>
              <div className="text-3xl font-bold">{streak.current_streak} days</div>
            </div>
          </div>

          <div className="text-right">
            <div className="text-xs opacity-75 mb-2">Personal Best</div>
            <div className="text-2xl font-bold">{streak.longest_streak} days</div>
          </div>
        </div>

        {/* Streak info */}
        <div className="mt-4 pt-4 border-t border-white border-opacity-20">
          <div className="grid grid-cols-2 gap-2">
            <div>
              <div className="text-xs opacity-75">Next Reward</div>
              <div className="font-bold">@ {nextMilestone} days</div>
            </div>
            <div>
              <div className="text-xs opacity-75">Bonus XP</div>
              <div className="font-bold text-yellow-100">+{bonus} XP</div>
            </div>
          </div>
        </div>

        {/* Progress bar */}
        {streak.current_streak < nextMilestone && (
          <div className="mt-3 bg-white bg-opacity-20 rounded-full h-2 overflow-hidden">
            <motion.div
              className="bg-white h-full rounded-full"
              initial={{ width: 0 }}
              animate={{ width: `${(streak.current_streak / nextMilestone) * 100}%` }}
              transition={{ duration: 0.5 }}
            />
          </div>
        )}

        {/* Milestone message */}
        {streak.current_streak >= nextMilestone && (
          <motion.div
            className="mt-3 text-center font-bold text-yellow-100 text-sm"
            animate={{ scale: [1, 1.05, 1] }}
            transition={{ duration: 2, repeat: Infinity }}
          >
            🎉 Milestone Reached! Bonus +{bonus} XP awarded!
          </motion.div>
        )}
      </div>
    </motion.div>
  );
};

export default StreakWidget;
