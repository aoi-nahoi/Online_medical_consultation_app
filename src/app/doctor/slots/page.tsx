'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Calendar, Clock, Edit, Trash2, Plus, ArrowLeft, CheckCircle, XCircle } from 'lucide-react'
import toast from 'react-hot-toast'

interface Slot {
  id: number
  startTime: string
  endTime: string
  status: 'open' | 'blocked'
  notes?: string
}

export default function SlotsPage() {
  const [slots, setSlots] = useState<Slot[]>([])
  const [loading, setLoading] = useState(true)
  const [editingSlot, setEditingSlot] = useState<Slot | null>(null)
  const router = useRouter()

  useEffect(() => {
    fetchSlots()
  }, [])

  const fetchSlots = async () => {
    try {
      const token = localStorage.getItem('token')
      if (!token) {
        router.push('/')
        return
      }

      const response = await fetch('/api/v1/doctors/me/slots', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })

      if (response.ok) {
        const data = await response.json()
        console.log('Fetched slots data:', data.slots)
        setSlots(data.slots || [])
      } else {
        toast.error('診療枠の取得に失敗しました')
      }
    } catch (error) {
      toast.error('診療枠の取得に失敗しました')
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteSlot = async (slotId: number) => {
    if (!confirm('この診療枠を削除しますか？')) return

    try {
      const token = localStorage.getItem('token')
      const response = await fetch(`/api/v1/doctors/me/slots/${slotId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })

      if (response.ok) {
        toast.success('診療枠を削除しました')
        fetchSlots()
      } else {
        toast.error('診療枠の削除に失敗しました')
      }
    } catch (error) {
      toast.error('診療枠の削除に失敗しました')
    }
  }

  const handleStatusToggle = async (slotId: number, currentStatus: string) => {
    const newStatus = currentStatus === 'open' ? 'blocked' : 'open'
    
    try {
      const token = localStorage.getItem('token')
      const response = await fetch(`/api/v1/doctors/me/slots/${slotId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          status: newStatus
        })
      })

      if (response.ok) {
        toast.success(`診療枠を${newStatus === 'open' ? '開放' : 'ブロック'}しました`)
        fetchSlots()
      } else {
        toast.error('診療枠の更新に失敗しました')
      }
    } catch (error) {
      toast.error('診療枠の更新に失敗しました')
    }
  }

  const formatDateTime = (dateTime: string) => {
    try {
      console.log('Formatting dateTime:', dateTime)
      
      // RFC3339形式の日時文字列をパース
      const date = new Date(dateTime)
      console.log('Parsed date:', date)
      
      if (isNaN(date.getTime())) {
        console.error('Invalid date:', dateTime)
        return '無効な日時'
      }
      
      const formatted = date.toLocaleString('ja-JP', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
      })
      
      console.log('Formatted result:', formatted)
      return formatted
    } catch (error) {
      console.error('Date formatting error:', error, dateTime)
      return '日時エラー'
    }
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'open':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800"><CheckCircle className="w-3 h-3 mr-1" />開放</span>
      case 'blocked':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800"><XCircle className="w-3 h-3 mr-1" />ブロック</span>
      default:
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">{status}</span>
    }
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
              <button
                onClick={() => router.back()}
                className="mr-4 p-2 text-gray-400 hover:text-gray-600"
              >
                <ArrowLeft className="w-5 h-5" />
              </button>
              <h1 className="text-xl font-semibold text-gray-900">診療枠管理</h1>
            </div>
            <button
              onClick={() => router.push('/doctor/slots/new')}
              className="btn-primary flex items-center"
            >
              <Plus className="w-4 h-4 mr-2" />
              新規作成
            </button>
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* 統計 */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-2 bg-blue-100 rounded-lg">
                <Calendar className="w-6 h-6 text-blue-600" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">総診療枠数</p>
                <p className="text-2xl font-semibold text-gray-900">{slots.length}</p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-2 bg-green-100 rounded-lg">
                <CheckCircle className="w-6 h-6 text-green-600" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">開放中</p>
                <p className="text-2xl font-semibold text-gray-900">
                  {slots.filter(s => s.status === 'open').length}
                </p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-2 bg-red-100 rounded-lg">
                <XCircle className="w-6 h-6 text-red-600" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">ブロック中</p>
                <p className="text-2xl font-semibold text-gray-900">
                  {slots.filter(s => s.status === 'blocked').length}
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* 診療枠一覧 */}
        <div className="bg-white rounded-lg shadow">
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-lg font-medium text-gray-900">診療枠一覧</h2>
          </div>
          <div className="divide-y divide-gray-200">
            {slots.length === 0 ? (
              <div className="px-6 py-8 text-center text-gray-500">
                <Calendar className="w-12 h-12 mx-auto mb-4 text-gray-300" />
                <p>診療枠がありません</p>
                <p className="text-sm">新規作成してください</p>
              </div>
            ) : (
              slots.map((slot) => (
                <div key={slot.id} className="px-6 py-4">
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <div className="flex items-center space-x-3 mb-2">
                        <h3 className="text-sm font-medium text-gray-900">
                          診療枠 #{slot.id}
                        </h3>
                        {getStatusBadge(slot.status)}
                      </div>
                      <div className="text-sm text-gray-500 space-y-1">
                        <p className="flex items-center">
                          <Clock className="w-4 h-4 mr-2" />
                          {formatDateTime(slot.startTime)} - {formatDateTime(slot.endTime)}
                        </p>
                        {slot.notes && (
                          <p className="text-gray-600">{slot.notes}</p>
                        )}
                      </div>
                    </div>
                    <div className="flex space-x-2">
                      <button
                        onClick={() => handleStatusToggle(slot.id, slot.status)}
                        className={`px-3 py-1 text-xs font-medium rounded-md ${
                          slot.status === 'open'
                            ? 'text-red-700 bg-red-100 hover:bg-red-200'
                            : 'text-green-700 bg-green-100 hover:bg-green-200'
                        }`}
                      >
                        {slot.status === 'open' ? 'ブロック' : '開放'}
                      </button>
                      <button
                        onClick={() => handleDeleteSlot(slot.id)}
                        className="px-3 py-1 text-xs font-medium text-red-700 bg-red-100 hover:bg-red-200 rounded-md flex items-center"
                      >
                        <Trash2 className="w-3 h-3 mr-1" />
                        削除
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
