'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Calendar, MessageSquare, Video, FileText, LogOut, Plus, Clock, CheckCircle, XCircle, Users } from 'lucide-react'
import toast from 'react-hot-toast'

interface Appointment {
  id: number
  patient: {
    name: string
  }
  status: 'pending' | 'confirmed' | 'cancelled' | 'completed'
  startTime: string
  endTime: string
  notes?: string
}

interface Slot {
  id: number
  startTime: string
  endTime: string
  status: 'open' | 'blocked'
}

interface User {
  id: number
  email: string
  role: string
  doctorProfile?: {
    name: string
    specialty: string
    licenseNumber: string
  }
}

export default function DoctorDashboard() {
  const [user, setUser] = useState<User | null>(null)
  const [appointments, setAppointments] = useState<Appointment[]>([])
  const [slots, setSlots] = useState<Slot[]>([])
  const [loading, setLoading] = useState(true)
  const router = useRouter()

  useEffect(() => {
    const token = localStorage.getItem('token')
    const userData = localStorage.getItem('user')

    if (!token || !userData) {
      router.push('/')
      return
    }

    try {
      const user = JSON.parse(userData)
      setUser(user)
      fetchData(token)
    } catch (error) {
      console.error('Error parsing user data:', error)
      router.push('/')
    }
  }, [router])

  const fetchData = async (token: string) => {
    try {
      const [appointmentsRes, slotsRes] = await Promise.all([
        fetch('/api/v1/doctors/me/appointments', {
          headers: { 'Authorization': `Bearer ${token}` },
        }),
        fetch('/api/v1/doctors/me/slots', {
          headers: { 'Authorization': `Bearer ${token}` },
        }),
      ])

      if (appointmentsRes.ok) {
        const data = await appointmentsRes.json()
        setAppointments(data.appointments || [])
      }

      if (slotsRes.ok) {
        const data = await slotsRes.json()
        setSlots(data.slots || [])
      }
    } catch (error) {
      toast.error('データの取得に失敗しました')
    } finally {
      setLoading(false)
    }
  }

  const handleLogout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    router.push('/')
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'pending':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800"><Clock className="w-3 h-3 mr-1" />保留中</span>
      case 'confirmed':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800"><CheckCircle className="w-3 h-3 mr-1" />確定</span>
      case 'cancelled':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800"><XCircle className="w-3 h-3 mr-1" />キャンセル</span>
      case 'completed':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">完了</span>
      default:
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">{status}</span>
    }
  }

  const formatDateTime = (dateTime: string) => {
    return new Date(dateTime).toLocaleString('ja-JP', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">読み込み中...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* ヘッダー */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center">
              <h1 className="text-xl font-semibold text-gray-900">医師ダッシュボード</h1>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-700">
                {user?.doctorProfile?.name || '医師'}先生
              </span>
              <span className="text-sm text-gray-500">
                {user?.doctorProfile?.specialty}
              </span>
              <button
                onClick={handleLogout}
                className="flex items-center text-sm text-gray-600 hover:text-gray-900"
              >
                <LogOut className="w-4 h-4 mr-1" />
                ログアウト
              </button>
            </div>
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* 統計カード */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-2 bg-blue-100 rounded-lg">
                <Users className="w-6 h-6 text-blue-600" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">総予約数</p>
                <p className="text-2xl font-semibold text-gray-900">{appointments.length}</p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-2 bg-yellow-100 rounded-lg">
                <Clock className="w-6 h-6 text-yellow-600" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">承認待ち</p>
                <p className="text-2xl font-semibold text-gray-900">
                  {appointments.filter(a => a.status === 'pending').length}
                </p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-2 bg-green-100 rounded-lg">
                <Calendar className="w-6 h-6 text-green-600" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">診療枠</p>
                <p className="text-2xl font-semibold text-gray-900">{slots.length}</p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-2 bg-purple-100 rounded-lg">
                <Video className="w-6 h-6 text-purple-600" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">今日の予約</p>
                <p className="text-2xl font-semibold text-gray-900">
                  {appointments.filter(a => {
                    const today = new Date().toDateString()
                    const appointmentDate = new Date(a.startTime).toDateString()
                    return appointmentDate === today && a.status === 'confirmed'
                  }).length}
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* アクションボタン */}
        <div className="mb-8">
          <div className="flex space-x-4">
            <button
              onClick={() => router.push('/doctor/slots/new')}
              className="btn-primary flex items-center"
            >
              <Plus className="w-4 h-4 mr-2" />
              診療枠作成
            </button>
            <button
              onClick={() => router.push('/doctor/slots')}
              className="btn-secondary flex items-center"
            >
              <Calendar className="w-4 h-4 mr-2" />
              枠管理
            </button>
            <button
              onClick={() => router.push('/doctor/appointments')}
              className="btn-secondary flex items-center"
            >
              <Users className="w-4 h-4 mr-2" />
              予約管理
            </button>
            <button
              onClick={() => router.push('/doctor/prescriptions')}
              className="btn-secondary flex items-center"
            >
              <FileText className="w-4 h-4 mr-2" />
              処方管理
            </button>
          </div>
        </div>

        {/* 承認待ちの予約 */}
        <div className="bg-white rounded-lg shadow mb-8">
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-lg font-medium text-gray-900">承認待ちの予約</h2>
          </div>
          <div className="divide-y divide-gray-200">
            {appointments.filter(a => a.status === 'pending').length === 0 ? (
              <div className="px-6 py-8 text-center text-gray-500">
                <Clock className="w-12 h-12 mx-auto mb-4 text-gray-300" />
                <p>承認待ちの予約がありません</p>
              </div>
            ) : (
              appointments
                .filter(a => a.status === 'pending')
                .slice(0, 5)
                .map((appointment) => (
                  <div key={appointment.id} className="px-6 py-4">
                    <div className="flex items-center justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-3">
                          <h3 className="text-sm font-medium text-gray-900">
                            {appointment.patient.name} 患者
                          </h3>
                          {getStatusBadge(appointment.status)}
                        </div>
                        <div className="mt-1 text-sm text-gray-500">
                          <p>{formatDateTime(appointment.startTime)} - {formatDateTime(appointment.endTime)}</p>
                          {appointment.notes && (
                            <p className="mt-1">{appointment.notes}</p>
                          )}
                        </div>
                      </div>
                      <div className="flex space-x-2">
                        <button
                          onClick={() => router.push(`/doctor/appointments/${appointment.id}/confirm`)}
                          className="btn-primary text-xs"
                        >
                          <CheckCircle className="w-3 h-3 mr-1" />
                          承認
                        </button>
                        <button
                          onClick={() => router.push(`/doctor/appointments/${appointment.id}/reject`)}
                          className="btn-secondary text-xs"
                        >
                          <XCircle className="w-3 h-3 mr-1" />
                          却下
                        </button>
                      </div>
                    </div>
                  </div>
                ))
            )}
          </div>
        </div>

        {/* 今日の予約 */}
        <div className="bg-white rounded-lg shadow">
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-lg font-medium text-gray-900">今日の予約</h2>
          </div>
          <div className="divide-y divide-gray-200">
            {appointments.filter(a => {
              const today = new Date().toDateString()
              const appointmentDate = new Date(a.startTime).toDateString()
              return appointmentDate === today && a.status === 'confirmed'
            }).length === 0 ? (
              <div className="px-6 py-8 text-center text-gray-500">
                <Calendar className="w-12 h-12 mx-auto mb-4 text-gray-300" />
                <p>今日の予約がありません</p>
              </div>
            ) : (
              appointments
                .filter(a => {
                  const today = new Date().toDateString()
                  const appointmentDate = new Date(a.startTime).toDateString()
                  return appointmentDate === today && a.status === 'confirmed'
                })
                .map((appointment) => (
                  <div key={appointment.id} className="px-6 py-4">
                    <div className="flex items-center justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-3">
                          <h3 className="text-sm font-medium text-gray-900">
                            {appointment.patient.name} 患者
                          </h3>
                          {getStatusBadge(appointment.status)}
                        </div>
                        <div className="mt-1 text-sm text-gray-500">
                          <p>{formatDateTime(appointment.startTime)} - {formatDateTime(appointment.endTime)}</p>
                          {appointment.notes && (
                            <p className="mt-1">{appointment.notes}</p>
                          )}
                        </div>
                      </div>
                      <div className="flex space-x-2">
                        <button
                          onClick={() => router.push(`/doctor/chat/${appointment.id}`)}
                          className="btn-secondary text-xs"
                        >
                          <MessageSquare className="w-3 h-3 mr-1" />
                          チャット
                        </button>
                        <button
                          onClick={() => router.push(`/doctor/video/${appointment.id}`)}
                          className="btn-primary text-xs"
                        >
                          <Video className="w-3 h-3 mr-1" />
                          通話開始
                        </button>
                      </div>
                    </div>
                  </div>
                ))
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
