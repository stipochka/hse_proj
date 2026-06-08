import { useEffect } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { initKeycloak } from '@/shared/lib/keycloak'
import { useAuthStore } from '@/app/store/authStore'
import MainLayout from '@/app/layout/MainLayout'
import DashboardPage from '@/pages/dashboard/ui/DashboardPage'
import ActivitiesPage from '@/pages/activities/ui/ActivitiesPage'
import ActivityEvaluationPage from '@/pages/activity-evaluation/ui/ActivityEvaluationPage'
import GroupStudentsPage from '@/pages/group-students/ui/GroupStudentsPage'
import ExportPage from '@/pages/export/ui/ExportPage'

function App() {
  const setUser = useAuthStore((state) => state.setUser)

  useEffect(() => {
    const initAuth = async () => {
      try {
        const user = await initKeycloak()
        if (user) {
          setUser(user)
        }
      } catch (error) {
        console.error('Failed to initialize Keycloak:', error)
      }
    }

    initAuth()
  }, [setUser])

  return (
    <MainLayout>
      <Routes>
        <Route path="/" element={<DashboardPage />} />
        <Route path="/activities" element={<ActivitiesPage />} />
        <Route path="/activities/:id/evaluate" element={<ActivityEvaluationPage />} />
        <Route path="/group-students" element={<GroupStudentsPage />} />
        <Route path="/export" element={<ExportPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </MainLayout>
  )
}

export default App

