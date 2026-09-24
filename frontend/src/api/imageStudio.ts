import { apiClient } from './client'

export interface ImageStudioGalleryEntry {
  id: string
  url: string
  prompt: string
  model: string
  size: string
  format: string
  created_at: number
  expires_at: number
}

export interface SaveImageStudioGalleryRequest {
  image_data_url: string
  prompt: string
  model: string
  size: string
  format: string
}

export async function listImageStudioGallery(): Promise<ImageStudioGalleryEntry[]> {
  const { data } = await apiClient.get<ImageStudioGalleryEntry[]>('/image-studio/gallery')
  return data
}

export async function saveImageStudioGallery(
  request: SaveImageStudioGalleryRequest,
): Promise<ImageStudioGalleryEntry> {
  const { data } = await apiClient.post<ImageStudioGalleryEntry>('/image-studio/gallery', request)
  return data
}

export const imageStudioGalleryAPI = {
  list: listImageStudioGallery,
  save: saveImageStudioGallery,
}

export default imageStudioGalleryAPI
