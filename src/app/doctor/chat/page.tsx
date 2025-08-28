'use client';

import { useState, useEffect, useRef } from 'react';
import { useParams } from 'next/navigation';

interface ChatMessage {
  id: number;
  content: string;
  sender_id: number;
  appointment_id: number;
  created_at: string;
  attachment_url?: string;
  attachment_type?: string;
  read_at?: string;
  sender: {
    name: string;
    role: string;
  };
}

interface Appointment {
  id: number;
  patient: {
    name: string;
    id: number;
  };
  start_time: string;
  end_time: string;
  status: string;
}

export default function DoctorChatPage() {
  const params = useParams();
  const appointmentId = params.appointmentId as string;
  
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [appointment, setAppointment] = useState<Appointment | null>(null);
  const [newMessage, setNewMessage] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [unreadCount, setUnreadCount] = useState(0);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (appointmentId) {
      loadAppointment();
      loadMessages();
      loadUnreadCount();
    }
  }, [appointmentId]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const loadAppointment = async () => {
    try {
      const response = await fetch(`/api/v1/appointments/${appointmentId}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });
      
      if (response.ok) {
        const data = await response.json();
        setAppointment(data.appointment);
      }
    } catch (error) {
      console.error('予約情報の読み込みに失敗しました:', error);
    }
  };

  const loadMessages = async () => {
    try {
      setIsLoading(true);
      const response = await fetch(`/api/v1/appointments/${appointmentId}/chat/messages`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });
      
      if (response.ok) {
        const data = await response.json();
        setMessages(data.messages || []);
      }
    } catch (error) {
      console.error('メッセージの読み込みに失敗しました:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const loadUnreadCount = async () => {
    try {
      const response = await fetch(`/api/v1/appointments/${appointmentId}/chat/unread-count`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });
      
      if (response.ok) {
        const data = await response.json();
        setUnreadCount(data.count || 0);
      }
    } catch (error) {
      console.error('未読数の取得に失敗しました:', error);
    }
  };

  const sendMessage = async () => {
    if (!newMessage.trim() && !selectedFile) return;

    try {
      const formData = new FormData();
      formData.append('content', newMessage);
      if (selectedFile) {
        formData.append('attachment', selectedFile);
      }

      const response = await fetch(`/api/v1/appointments/${appointmentId}/chat/messages`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: formData,
      });

      if (response.ok) {
        setNewMessage('');
        setSelectedFile(null);
        await loadMessages();
        await loadUnreadCount();
      }
    } catch (error) {
      console.error('メッセージの送信に失敗しました:', error);
    }
  };

  const markAsRead = async () => {
    try {
      await fetch(`/api/v1/appointments/${appointmentId}/chat/read`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });
      
      await loadUnreadCount();
    } catch (error) {
      console.error('既読の設定に失敗しました:', error);
    }
  };

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      if (file.size > 10 * 1024 * 1024) {
        alert('ファイルサイズは10MB以下にしてください。');
        return;
      }
      setSelectedFile(file);
    }
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const formatTime = (dateString: string) => {
    return new Date(dateString).toLocaleString('ja-JP');
  };

  const isCurrentUser = (senderId: number) => {
    // 実際の実装では、現在のユーザーIDと比較
    return true; // 医師は常に送信者として表示
  };

  if (!appointment) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-4xl mx-auto p-6">
        {/* ヘッダー */}
        <div className="bg-white rounded-lg shadow-sm p-6 mb-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold text-gray-900">チャット</h1>
              <p className="text-gray-600">
                患者: {appointment.patient.name} | 予約ID: {appointmentId}
              </p>
              <p className="text-sm text-gray-500">
                予約時間: {new Date(appointment.start_time).toLocaleString('ja-JP')}
              </p>
            </div>
            <div className="text-right">
              <div className="text-sm text-gray-600">未読メッセージ</div>
              <div className="text-2xl font-bold text-blue-600">{unreadCount}</div>
              {unreadCount > 0 && (
                <button
                  onClick={markAsRead}
                  className="mt-2 px-3 py-1 bg-blue-600 text-white rounded text-sm hover:bg-blue-700"
                >
                  全て既読にする
                </button>
              )}
            </div>
          </div>
        </div>

        {/* チャットエリア */}
        <div className="bg-white rounded-lg shadow-sm">
          {/* メッセージ表示エリア */}
          <div className="h-96 overflow-y-auto p-6 border-b">
            {isLoading ? (
              <div className="flex justify-center items-center h-full">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
              </div>
            ) : messages.length === 0 ? (
              <div className="text-center text-gray-500 py-8">
                まだメッセージがありません
              </div>
            ) : (
              <div className="space-y-4">
                {messages.map((message) => (
                  <div
                    key={message.id}
                    className={`flex ${isCurrentUser(message.sender_id) ? 'justify-end' : 'justify-start'}`}
                  >
                    <div
                      className={`max-w-xs lg:max-w-md px-4 py-2 rounded-lg ${
                        isCurrentUser(message.sender_id)
                          ? 'bg-blue-600 text-white'
                          : 'bg-gray-200 text-gray-900'
                      }`}
                    >
                      <div className="text-sm font-medium mb-1">
                        {message.sender.name} ({message.sender.role})
                      </div>
                      <div className="mb-2">{message.content}</div>
                      
                      {/* 添付ファイル */}
                      {message.attachment_url && (
                        <div className="mt-2">
                          {message.attachment_type?.startsWith('image/') ? (
                            <img
                              src={message.attachment_url}
                              alt="添付画像"
                              className="max-w-full h-auto rounded"
                            />
                          ) : (
                            <a
                              href={message.attachment_url}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="text-blue-300 hover:text-blue-200 underline"
                            >
                              📎 添付ファイルを表示
                            </a>
                          )}
                        </div>
                      )}
                      
                      <div className="text-xs opacity-75 mt-1">
                        {formatTime(message.created_at)}
                        {!message.read_at && !isCurrentUser(message.sender_id) && (
                          <span className="ml-2 text-yellow-300">未読</span>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
                <div ref={messagesEndRef} />
              </div>
            )}
          </div>

          {/* メッセージ入力エリア */}
          <div className="p-6">
            <div className="flex space-x-4">
              <div className="flex-1">
                <textarea
                  value={newMessage}
                  onChange={(e) => setNewMessage(e.target.value)}
                  placeholder="メッセージを入力してください..."
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
                  rows={3}
                />
              </div>
              <div className="flex flex-col space-y-2">
                <button
                  onClick={sendMessage}
                  disabled={!newMessage.trim() && !selectedFile}
                  className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  送信
                </button>
                <label className="px-6 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 cursor-pointer text-center">
                  📎 ファイル
                  <input
                    type="file"
                    accept="image/*,.pdf"
                    onChange={handleFileSelect}
                    className="hidden"
                  />
                </label>
              </div>
            </div>
            
            {/* 選択されたファイル表示 */}
            {selectedFile && (
              <div className="mt-4 p-3 bg-blue-50 rounded-lg">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-blue-800">
                    📎 {selectedFile.name} ({(selectedFile.size / 1024 / 1024).toFixed(2)}MB)
                  </span>
                  <button
                    onClick={() => setSelectedFile(null)}
                    className="text-blue-600 hover:text-blue-800"
                  >
                    ✕
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
