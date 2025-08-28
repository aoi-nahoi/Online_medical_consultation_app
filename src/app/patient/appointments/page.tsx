'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'

interface Appointment {
  id: number
  doctor: {
    name: string
    specialty: string
  }
  status: string
  start_time: string
  end_time: string
  notes: string
}

export default function PatientAppointments() {
  const [appointments, setAppointments] = useState<Appointment[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const router = useRouter()

  useEffect(() => {
    fetchAppointments()
  }, [])

  const fetchAppointments = async () => {
    try {
      const token = localStorage.getItem('token')
      if (!token) {
        router.push('/login')
        return
      }

      const response = await fetch('/api/v1/patients/appointments', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })

      if (!response.ok) {
        throw new Error('予約の取得に失敗しました')
      }

      const data = await response.json()
      setAppointments(data.appointments)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'エラーが発生しました')
    } finally {
      setLoading(false)
    }
  }

  const cancelAppointment = async (appointmentId: number) => {
    if (!confirm('この予約をキャンセルしますか？')) {
      return
    }

    try {
      const token = localStorage.getItem('token')
      const response = await fetch(`/api/v1/patients/appointments/${appointmentId}/cancel`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      })

      if (!response.ok) {
        throw new Error('予約のキャンセルに失敗しました')
      }

      // 予約一覧を再取得
      fetchAppointments()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'エラーが発生しました')
    }
  }

  const getStatusBadge = (status: string) => {
    const statusMap = {
      pending: { label: '保留中', color: 'bg-yellow-100 text-yellow-800' },
      confirmed: { label: '確定済み', color: 'bg-green-100 text-green-800' },
      cancelled: { label: 'キャンセル', color: 'bg-red-100 text-red-800' },
      completed: { label: '完了', color: 'bg-blue-100 text-blue-800' }
    }
    const statusInfo = statusMap[status as keyof typeof statusMap] || { label: status, color: 'bg-gray-100 text-gray-800' }
    
    return (
      <span className={`px-2 py-1 text-xs font-medium rounded-full ${statusInfo.color}`}>
        {statusInfo.label}
      </span>
    )
  }

  const formatDateTime = (dateTime: string) => {
    return new Date(dateTime).toLocaleString('ja-JP', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    })
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">読み込み中...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="text-red-600 text-xl mb-4">エラー</div>
          <p className="text-gray-600 mb-4">{error}</p>
          <button
            onClick={fetchAppointments}
            className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
          >
            再試行
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">予約管理</h1>
          <p className="mt-2 text-gray-600">あなたの予約一覧と管理を行えます</p>
        </div>

        <div className="bg-white shadow rounded-lg">
          <div className="px-6 py-4 border-b border-gray-200">
            <div className="flex justify-between items-center">
              <h2 className="text-lg font-medium text-gray-900">予約一覧</h2>
              <button
                onClick={() => router.push('/patient/appointments/new')}
                className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors"
              >
                新規予約
              </button>
            </div>
          </div>

          {appointments.length === 0 ? (
            <div className="px-6 py-12 text-center">
              <div className="text-gray-400 text-6xl mb-4">📅</div>
              <h3 className="text-lg font-medium text-gray-900 mb-2">予約がありません</h3>
              <p className="text-gray-500">新しい予約を作成して、医師との相談を始めましょう</p>
              <button
                onClick={() => router.push('/patient/appointments/new')}
                className="mt-4 bg-blue-600 text-white px-6 py-2 rounded-md hover:bg-blue-700 transition-colors"
              >
                予約を作成
              </button>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      医師
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      診療科
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      日時
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      ステータス
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      メモ
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      操作
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {appointments.map((appointment) => (
                    <tr key={appointment.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm font-medium text-gray-900">
                          {appointment.doctor.name}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm text-gray-500">
                          {appointment.doctor.specialty}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm text-gray-900">
                          {formatDateTime(appointment.start_time)}
                        </div>
                        <div className="text-sm text-gray-500">
                          〜 {formatDateTime(appointment.end_time)}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {getStatusBadge(appointment.status)}
                      </td>
                      <td className="px-6 py-4">
                        <div className="text-sm text-gray-900 max-w-xs truncate">
                          {appointment.notes || 'なし'}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                        <div className="flex space-x-2">
                          <button
                            onClick={() => router.push(`/patient/appointments/${appointment.id}`)}
                            className="text-blue-600 hover:text-blue-900"
                          >
                            詳細
                          </button>
                          {appointment.status === 'confirmed' && (
                            <>
                              <button
                                onClick={() => router.push(`/patient/appointments/${appointment.id}/chat`)}
                                className="text-green-600 hover:text-green-900"
                              >
                                チャット
                              </button>
                              <button
                                onClick={() => router.push(`/patient/appointments/${appointment.id}/video`)}
                                className="text-purple-600 hover:text-purple-900"
                              >
                                ビデオ通話
                              </button>
                            </>
                          )}
                          {appointment.status === 'pending' && (
                            <button
                              onClick={() => cancelAppointment(appointment.id)}
                              className="text-red-600 hover:text-red-900"
                            >
                              キャンセル
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
