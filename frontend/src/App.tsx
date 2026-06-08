import { Routes, Route, Navigate } from 'react-router-dom'
import MainLayout from '@/app/layout/MainLayout'
import DashboardPage from '@/pages/dashboard/ui/DashboardPage'
import ActivitiesPage from '@/pages/activities/ui/ActivitiesPage'
import ActivityEvaluationPage from '@/pages/activity-evaluation/ui/ActivityEvaluationPage'
import GroupStudentsPage from '@/pages/group-students/ui/GroupStudentsPage'
import ExportPage from '@/pages/export/ui/ExportPage'
import SubmitActivityPage from '@/pages/submit-activity/ui/SubmitActivityPage'
import { useAuthStore } from '@/app/store/authStore'

const isAdmin = (roles: string[]) =>
  roles.includes('group_admin') || roles.includes('super_admin')

function App() {
  const user = useAuthStore((s) => s.user)
  const admin = isAdmin(user?.roles ?? [])

  return (
    <MainLayout>
      <Routes>
        <Route path="/" element={<DashboardPage />} />
        {/* /submit только для студентов */}
        <Route path="/submit" element={admin ? <Navigate to="/" replace /> : <SubmitActivityPage />} />
        <Route path="/activities" element={<ActivitiesPage />} />
        <Route path="/activities/:id/evaluate" element={<ActivityEvaluationPage />} />
        {/* /group-students и evaluate только для админов */}
        <Route path="/group-students" element={admin ? <GroupStudentsPage /> : <Navigate to="/" replace />} />
        <Route path="/export" element={<ExportPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </MainLayout>
  )
}

export default App
