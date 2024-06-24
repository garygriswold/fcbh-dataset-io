

# https://abdeladim-s.github.io/easymms/#easymms.models.alignment.AlignmentModel

# https://github.com/abdeladim-s/easymms/blob/main/README.md


# /usr/bin/pip3 install easymms

# /usr/bin/pip3 install -U --pre torchaudio --index-url https://download.pytorch.org/whl/nightly/cu118

# /usr/bin/pip3 install tensorboardX
import os
from easymms.models.asr import ASRModel

data = os.environ['FCBH_DATASET_DB']

model = data + '/easy_mms/models/mms1b_all.pt'
print("model", model)
file1 = data + '/download/ENGWEB/ENGWEBN2DA-mp3-64/B25___01_3John_______ENGWEBN2DA.mp3'
print("file", file1)
asr = ASRModel(model=model)
files = [file1]
transcriptions = asr.transcribe(files, lang='eng', align=False)
for i, transcription in enumerate(transcriptions):
    #print(f">>> file {files[i]}")
    print(transcription)


# /usr/bin/pip3 install --upgrade fairseq

# /usr/bin/pip3 install --upgrade pytorch